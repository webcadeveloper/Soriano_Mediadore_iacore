package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient     *mongo.Client
	MongoCacheDB    *mongo.Database    // DB1: Query caching & búsquedas rápidas
	MongoAnalyticsDB *mongo.Database   // DB2: Reports & métricas
	MongoSessionsDB *mongo.Database    // DB3: Sesiones de usuario & conversaciones bot
)

// InitMongo conecta a MongoDB (3 bases de datos)
func InitMongo() error {
	mongoURI := fmt.Sprintf("mongodb://%s:%s",
		os.Getenv("MONGO_HOST"),
		os.Getenv("MONGO_PORT"),
	)

	// Configurar opciones de conexión
	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetMaxPoolSize(50)
	clientOptions.SetMinPoolSize(10)

	// Conectar
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	MongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error conectando a MongoDB: %w", err)
	}

	// Verificar conexión
	if err = MongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error haciendo ping a MongoDB: %w", err)
	}

	// Inicializar las 3 bases de datos
	MongoCacheDB = MongoClient.Database(os.Getenv("MONGO_DB_CACHE"))
	MongoAnalyticsDB = MongoClient.Database(os.Getenv("MONGO_DB_ANALYTICS"))
	MongoSessionsDB = MongoClient.Database(os.Getenv("MONGO_DB_SESSIONS"))

	log.Println("✅ Conectado a MongoDB (3 bases de datos)")
	log.Printf("   - DB Cache: %s", os.Getenv("MONGO_DB_CACHE"))
	log.Printf("   - DB Analytics: %s", os.Getenv("MONGO_DB_ANALYTICS"))
	log.Printf("   - DB Sessions: %s", os.Getenv("MONGO_DB_SESSIONS"))

	return nil
}

// GuardarBusquedaCache guarda resultados de búsqueda en MongoDB cache
func GuardarBusquedaCache(query string, resultados interface{}) error {
	// Si MongoDB no está disponible, no hacer nada (no es crítico)
	if MongoCacheDB == nil {
		return nil
	}

	collection := MongoCacheDB.Collection("busquedas")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := map[string]interface{}{
		"query":      query,
		"resultados": resultados,
		"timestamp":  time.Now(),
	}

	_, err := collection.InsertOne(ctx, doc)
	return err
}

// GuardarSesionBot guarda una sesión de conversación con bot
func GuardarSesionBot(sessionID, botID string, mensaje map[string]interface{}) error {
	// Si MongoDB no está disponible, no hacer nada (no es crítico)
	if MongoSessionsDB == nil {
		return nil
	}

	collection := MongoSessionsDB.Collection("conversaciones")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := map[string]interface{}{
		"session_id": sessionID,
		"bot_id":     botID,
		"mensaje":    mensaje,
		"timestamp":  time.Now(),
	}

	_, err := collection.InsertOne(ctx, doc)
	return err
}

// GuardarMetrica guarda métricas de analytics
func GuardarMetrica(tipoMetrica string, datos map[string]interface{}) error {
	// Si MongoDB no está disponible, no hacer nada (no es crítico)
	if MongoAnalyticsDB == nil {
		return nil
	}

	collection := MongoAnalyticsDB.Collection("metricas")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := map[string]interface{}{
		"tipo":      tipoMetrica,
		"datos":     datos,
		"timestamp": time.Now(),
	}

	_, err := collection.InsertOne(ctx, doc)
	return err
}
