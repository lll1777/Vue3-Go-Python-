from flask import Blueprint, request, jsonify
from services.prediction_service import PredictionService

prediction_bp = Blueprint('prediction', __name__)
prediction_service = PredictionService()


@prediction_bp.route('/traffic', methods=['GET'])
def get_traffic_prediction():
    try:
        hours = request.args.get('hours', 24, type=int)
        prediction = prediction_service.predict_traffic(hours)
        
        return jsonify({
            'success': True,
            'data': prediction
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@prediction_bp.route('/parking', methods=['GET'])
def get_parking_prediction():
    try:
        zone = request.args.get('zone', None)
        prediction = prediction_service.predict_occupancy(zone)
        
        return jsonify({
            'success': True,
            'data': prediction
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@prediction_bp.route('/peak', methods=['GET'])
def get_peak_prediction():
    try:
        peak_info = prediction_service.predict_peak_hours()
        
        return jsonify({
            'success': True,
            'data': peak_info
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@prediction_bp.route('/vacancy', methods=['GET'])
def get_vacancy_prediction():
    try:
        time = request.args.get('time', None)
        vacancy = prediction_service.predict_vacancy(time)
        
        return jsonify({
            'success': True,
            'data': vacancy
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@prediction_bp.route('/train', methods=['POST'])
def train_model():
    try:
        data = request.get_json() if request.is_json else {}
        
        result = prediction_service.train_model(data)
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@prediction_bp.route('/model/status', methods=['GET'])
def get_model_status():
    try:
        status = prediction_service.get_model_status()
        
        return jsonify({
            'success': True,
            'data': status
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500
