# genre_analysis_service/models.py (ИСПРАВЛЕННАЯ ВЕРСИЯ)

from django.db import models

class Genre(models.Model):
    genre_id = models.IntegerField(primary_key=True)
    genre_name = models.CharField(max_length=100)
    genre_keywords = models.TextField(default="") 

    class Meta:
        db_table = 'genres' 
        managed = False
        
class AnalysisRequest(models.Model):
    analysis_request_id = models.IntegerField(primary_key=True)
    text_to_analyse = models.TextField()

    class Meta:
        db_table = 'analysis_requests'
        managed = False
        
# 🔥 ИСПРАВЛЕННАЯ МОДЕЛЬ AnalysisGenre
class AnalysisGenre(models.Model):
    
    # 1. Мы объявляем поля, составляющие составной ключ, 
    #    но primary_key=True ставим только на ОДНО из них, чтобы
    #    избежать E026 и ошибки "id does not exist".
    
    analysis_request = models.ForeignKey(
        'AnalysisRequest', 
        on_delete=models.DO_NOTHING, 
        db_column='analysis_request_id', 
        related_name='analysis_genres',
        primary_key=True # 🔥 Удовлетворяем требованию Django о наличии PK
    )
    
    genre = models.ForeignKey(
        'Genre', 
        on_delete=models.DO_NOTHING, 
        db_column='genre_id', 
        related_name='analysis_genres',
        # 🔥 Здесь НЕ ставим primary_key=True
    )
    
    # 🔥 УДАЛЯЕМ строку: id = models.AutoField(primary_key=True)

    class Meta:
        db_table = 'analysis_genres'
        managed = False
        # Объявляем составной ключ через unique_together
        unique_together = (('analysis_request', 'genre'),)