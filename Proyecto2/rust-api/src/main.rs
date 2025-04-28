use actix_web::{web, App, HttpResponse, HttpServer, Responder, middleware::Logger};
use serde::{Deserialize, Serialize};
use std::env;
use std::sync::Arc;
use reqwest::Client;
use std::time::Duration;
use log::{info, error};

#[derive(Debug, Serialize, Deserialize, Clone)]
struct WeatherData {
    descripcion: String,
    Pais: String,
    Clima: String,
}

#[derive(Clone)]
struct AppState {
    client: Client,
    go_api_url: String,
}

async fn forward_to_go(
    weather_data: web::Json<Vec<WeatherData>>,
    state: web::Data<Arc<AppState>>,
) -> impl Responder {
    info!("Received {} weather data entries", weather_data.len());
    
    match state.client.post(&state.go_api_url)
        .json(&weather_data.into_inner())
        .send()
        .await {
            Ok(response) => {
                if response.status().is_success() {
                    info!("Successfully forwarded data to Go API");
                    HttpResponse::Ok().body("Data processed successfully")
                } else {
                    let status = response.status();
                    error!("Go API returned error status: {}", status);
                    HttpResponse::InternalServerError().body(format!("Go API error: {}", status))
                }
            },
            Err(e) => {
                error!("Failed to forward data to Go API: {}", e);
                HttpResponse::InternalServerError().body(format!("Failed to forward data: {}", e))
            }
        }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::default().default_filter_or("info"));
    
    let go_api_url = env::var("GO_API_URL").unwrap_or_else(|_| "http://go-api-service:8080/process".to_string());
    
    let client = Client::builder()
        .timeout(Duration::from_secs(30))
        .build()
        .expect("Failed to create HTTP client");
    
    let app_state = Arc::new(AppState {
        client,
        go_api_url,
    });
    
    let port = env::var("PORT").unwrap_or_else(|_| "8000".to_string());
    let addr = format!("0.0.0.0:{}", port);
    
    info!("Starting Rust API server on {}", addr);
    info!("Forwarding to Go API at {}", go_api_url);
    
    HttpServer::new(move || {
        App::new()
            .wrap(Logger::default())
            .app_data(web::Data::new(app_state.clone()))
            .route("/input", web::post().to(forward_to_go))
            .route("/health", web::get().to(|| async { HttpResponse::Ok().body("Healthy") }))
    })
    .bind(addr)?
    .run()
    .await
}