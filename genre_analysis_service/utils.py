# genre_analysis_service/utils.py

import re

def calculate_genre_probability(text_to_analyse: str, genre_keywords: str) -> int:
    """
    Рассчитывает процент вероятности принадлежности текста к жанру.
    Возвращает целое число от 0 до 100.
    """
    if not text_to_analyse or not genre_keywords:
        return 0

    text = text_to_analyse.lower()
    
    # 1. Подготовка ключевых слов
    # Разделение ключевых слов: по запятой, с очисткой пробелов
    keywords = [kw.strip() for kw in genre_keywords.lower().split(',') if kw.strip()]
    
    match_count = 0
    for kw in keywords:
        # 2. Подсчет совпадений
        # Проверяем, содержится ли ключевое слово в тексте
        if kw in text:
            match_count += 1
            
    # 3. Подсчет общего количества слов в тексте (для нормировки)
    # Используем регулярное выражение для поиска слов
    total_words = len(re.findall(r'\b\w+\b', text_to_analyse))

    if total_words == 0:
        return 0

    # 4. Расчет процента
    # Мы нормируем совпадения на общее количество слов (как это было в Go)
    probability = (match_count / total_words) * 100
    
    # 5. Возвращаем округленное целое число
    return min(100, round(probability))