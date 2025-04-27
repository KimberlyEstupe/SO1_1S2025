// api_rest/src/main.rs
use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use reqwest::Client;
use std::sync::Arc;
use log::{info, error};

#[derive(Debug, Serialize, Deserialize, Clone)]
struct WeatherTweet {
    descripcion: String,
    Pais: String,
    Clima: String,
}

async fn forward_to_go_api(
    weather_tweets: web::Json<Vec<WeatherTweet>>,
    client: web::Data<Client>,
) -> impl Responder {
    info!("Recibida petición con {} tweets", weather_tweets.len());
    
    let go_api_url = std::env::var("GO_API_URL")
        .unwrap_or_else(|_| "http://go-api-service.weather-app.svc.cluster.local/process".to_string());
    
    match client.post(&go_api_url)
        .json(&weather_tweets.into_inner())
        .send()
        .await {
            Ok(response) => {
                match response.status().is_success() {
                    true => {
                        info!("Petición enviada correctamente al API de Go");
                        HttpResponse::Ok().body("Tweets procesados correctamente")
                    },
                    false => {
                        error!("Error del servidor Go: {}", response.status());
                        HttpResponse::InternalServerError().body("Error al procesar tweets")
                    }
                }
            },
            Err(e) => {
                error!("Error al enviar petición al API de Go: {}", e);
                HttpResponse::InternalServerError().body(format!("Error: {}", e))
            }
        }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::default().default_filter_or("info"));
    
    info!("Iniciando API REST en Rust...");
    
    // Crear cliente HTTP para llamadas al API de Go
    let client = Client::new();
    let client_data = web::Data::new(client);
    
    // Configurar y lanzar el servidor HTTP
    HttpServer::new(move || {
        App::new()
            .app_data(client_data.clone())
            .route("/input", web::post().to(forward_to_go_api))
            .route("/health", web::get().to(|| async { HttpResponse::Ok().body("Healthy") }))
    })
    .bind("0.0.0.0:8000")?
    .workers(num_cpus::get() * 2) // Usar múltiples workers para manejar concurrencia
    .run()
    .await
}