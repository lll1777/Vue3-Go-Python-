from flask import Flask
from flask_cors import CORS
from config import config


def create_app(config_name='default'):
    app = Flask(__name__)
    app.config.from_object(config[config_name])
    
    CORS(app)
    
    from routes.prediction import prediction_bp
    from routes.plate_recognition import plate_bp
    from routes.anomaly import anomaly_bp
    
    app.register_blueprint(prediction_bp, url_prefix='/api/prediction')
    app.register_blueprint(plate_bp, url_prefix='/api/plate')
    app.register_blueprint(anomaly_bp, url_prefix='/api/anomaly')
    
    @app.route('/api/health')
    def health():
        return {
            'status': 'ok',
            'message': 'Python parking service is running'
        }
    
    return app


if __name__ == '__main__':
    app = create_app('development')
    app.run(host='0.0.0.0', port=5000, debug=True)
