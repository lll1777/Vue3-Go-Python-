import os
from datetime import timedelta


class Config:
    SECRET_KEY = os.environ.get('SECRET_KEY') or 'parking-python-secret-key'
    
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL') or \
        'sqlite:///parking_python.db'
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    
    GO_BACKEND_URL = os.environ.get('GO_BACKEND_URL') or 'http://localhost:8080'
    
    MODEL_PATH = os.environ.get('MODEL_PATH') or 'models/trained'
    
    PLATE_RECOGNITION_CONFIDENCE = float(os.environ.get('PLATE_RECOGNITION_CONFIDENCE', 0.95))
    
    PREDICTION_INTERVAL = int(os.environ.get('PREDICTION_INTERVAL', 3600))
    
    PERMANENT_SESSION_LIFETIME = timedelta(days=7)


class DevelopmentConfig(Config):
    DEBUG = True


class ProductionConfig(Config):
    DEBUG = False


config = {
    'development': DevelopmentConfig,
    'production': ProductionConfig,
    'default': DevelopmentConfig
}
