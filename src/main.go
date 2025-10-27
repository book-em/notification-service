package main

import (
	"bookem-notification-service/client/userclient"
	internal "bookem-notification-service/internal"
	"bookem-notification-service/util"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var (
	server      *gin.Engine
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
)

func connectToMongo() *mongo.Database {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("MongoDB not reachable: %v", err)
	}

	log.Printf("Connected to MongoDB at %s", mongoURI)
	mongoClient = client
	mongoDB = client.Database(dbName)
	return mongoDB
}

func healthHandler(ctx *gin.Context) {
	ctxMongo, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mongoClient.Ping(ctxMongo, readpref.Primary()); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "MongoDB not reachable"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func main() {
	ctx := context.Background()
	shutdown := util.TEL.Init(
		ctx,
		os.Getenv("SERVICE_NAME"),
		os.Getenv("DEPLOYMENT_ENV"),
	)
	defer shutdown(ctx)

	db := connectToMongo()
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	server = gin.Default()

	server.Use(internal.PrometheusMiddleware())
	server.Use(util.TEL.GetLoggingMiddleware())
	server.Use(otelgin.Middleware(os.Getenv("SERVICE_NAME")))
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost", "http://bookem.local"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server.GET("/metrics", gin.WrapH(promhttp.Handler()))
	server.GET("/healthz", healthHandler)

	userClient := userclient.NewUserClient()

	notificationRepo := internal.NewRepository(db)
	service := internal.NewService(notificationRepo, userClient)
	handler := internal.NewHandler(service)
	route := *internal.NewRoute(handler)

	rg := server.Group("/api")
	route.Route(rg)

	server.Run()
}
