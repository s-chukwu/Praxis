package config

import "os"

type Config struct {
	Port                   string
	CORSAllowedOrigin      string
	AIServiceURL           string
	AIServiceSharedSecret  string
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
}

func Load() Config {
	return Config{
		Port:                   getEnv("PORT", "8080"),
		CORSAllowedOrigin:      getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
		AIServiceURL:           getEnv("AI_SERVICE_URL", "http://ai:8000"),
		AIServiceSharedSecret:  os.Getenv("AI_SERVICE_SHARED_SECRET"),
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:        os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}