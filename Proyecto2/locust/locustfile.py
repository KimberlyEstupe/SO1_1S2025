from locust import HttpUser, task, between
import json
import random

# Weather options
weather_options = ["Lluvioso", "Nublado", "Soleado"]
countries = ["GT", "MX", "US", "CA", "BR", "AR", "CO", "PE", "CL", "ES"]
descriptions = {
    "Lluvioso": ["Esta lloviendo", "Lluvia fuerte", "Llovizna ligera"],
    "Nublado": ["Esta nublado", "Cielo gris", "Nubes densas"],
    "Soleado": ["Esta soleado", "Sol radiante", "Día despejado"]
}

class WeatherUser(HttpUser):
    wait_time = between(0.1, 0.5)
    
    def generate_weather_data(self, size=10000):
        data = []
        for _ in range(size):
            weather = random.choice(weather_options)
            country = random.choice(countries)
            desc = random.choice(descriptions[weather])
            
            data.append({
                "descripcion": desc,
                "Pais": country,
                "Clima": weather
            })
        return data
    
    @task
    def post_weather(self):
        payload = self.generate_weather_data()
        self.client.post("/input", json=payload)