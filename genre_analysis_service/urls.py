from django.urls import path
from . import views

urlpatterns = [
    path('calculate-text-genre-probability', views.start_analysis_process, name='start_analysis'),
]