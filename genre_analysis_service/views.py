# genre_analysis_service/views.py (ИСКЛЮЧИТЕЛЬНО РАСЧЁТ ВЕРОЯТНОСТИ)

import json
import time
import requests
import threading
import logging 
import random 
from django.http import HttpResponse, JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.conf import settings
from .utils import calculate_genre_probability 
# 🔥 Импортируем обе модели для доступа к данным
from .models import AnalysisGenre, AnalysisRequest 

logger = logging.getLogger(__name__)

# --- Вспомогательная функция для Callback'а ---
# Здесь мы передаем SECRET_KEY и используем PUT-метод (корректно)
def send_callback(analysis_request_id: int, analysis_genre_data: list):
    callback_url = settings.GO_CALLBACK_URL
    callback_data = {
        "analysis_request_id": analysis_request_id,
        "secret_key": settings.INTERNAL_SECRET_KEY, 
        "analysis_genre_data": analysis_genre_data,
    }

    try:
        response = requests.put(callback_url, json=callback_data)
        response.raise_for_status()
        logger.info(f"Заявка {analysis_request_id}: Callback успешно отправлен, статус: {response.status_code}")
    except requests.exceptions.RequestException as e:
        logger.error(f"Заявка {analysis_request_id}: Ошибка Callback'а к Go-сервису: {e}")
        pass


# --- Основная логика анализа (получает все из БД) ---

def perform_analysis_and_callback(analysis_request_id: int): # 🔥 УБРАЛИ text_to_analyse из параметров
    logger.info(f"Заявка {analysis_request_id}: Начато асинхронное выполнение. Получение данных из БД...")

    # 1. 🔥 ПОЛУЧАЕМ ТЕКСТ ДЛЯ АНАЛИЗА ИЗ БД
    try:
        analysis_request = AnalysisRequest.objects.get(pk=analysis_request_id)
        text_to_analyse = analysis_request.text_to_analyse 
        if not text_to_analyse:
            logger.error(f"Заявка {analysis_request_id}: Текст для анализа пуст.")
            return send_callback(analysis_request_id, [])
    except AnalysisRequest.DoesNotExist:
        logger.error(f"Заявка {analysis_request_id}: Заявка не найдена в БД.")
        return 
    except Exception as e:
        logger.error(f"Заявка {analysis_request_id}: Ошибка получения заявки из БД: {e}")
        return


    delay = random.randint(5, 10)
    time.sleep(delay)
    logger.info(f"Заявка {analysis_request_id}: Задержка {delay}с завершена, начинаю расчет.")
    
    analysis_genre_data = []
    
    try:
        # 2. 🔥 ПОЛУЧАЕМ ЖАНРЫ ДЛЯ АНАЛИЗА ИЗ БД (Остается так же)
        analysis_genres = AnalysisGenre.objects.filter(analysis_request_id=analysis_request_id).select_related('genre')

        if not analysis_genres.exists():
             logger.warning(f"Заявка {analysis_request_id}: Жанры для анализа не найдены.")
             send_callback(analysis_request_id, [])
             return 

        for ag in analysis_genres:
            keywords = ag.genre.genre_keywords  
            probability = 0
            if keywords:
                 # 3. 🔥 ИСПОЛЬЗУЕМ ПОЛУЧЕННЫЙ ИЗ БД ТЕКСТ
                 probability = calculate_genre_probability(text_to_analyse, keywords) 
            
            analysis_genre_data.append({
                "genre_id": ag.genre_id,
                "probability_percent": probability, 
            })
            
    except Exception as e:
        logger.error(f"Заявка {analysis_request_id}: Фатальная ошибка при расчете или доступе к БД: {e}")
        return
        
    send_callback(analysis_request_id, analysis_genre_data) 

@csrf_exempt
def start_analysis_process(request):
    """
    POST /calculate-text-genre-probability
    Запускает асинхронный расчет. Возвращает 204 No Content.
    """
    if request.method != 'POST':
        return JsonResponse({'message': 'Method not allowed'}, status=405)

    try:
        data = json.loads(request.body)
        analysis_request_id = data.get('analysis_request_id')
        
        # 🔥 УБРАЛИ ПРОВЕРКУ text_to_analyse
    except json.JSONDecodeError:
        return JsonResponse({'message': 'Invalid JSON'}, status=400)

    if not analysis_request_id:
         return JsonResponse({'message': 'Missing analysis_request_id'}, status=400)
         
    # 🔥 В perform_analysis_and_callback передаем только ID
    threading.Thread(target=perform_analysis_and_callback, args=(analysis_request_id,)).start() 
    
    return HttpResponse(status=204)