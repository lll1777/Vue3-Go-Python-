from flask import Blueprint, request, jsonify
from services.anomaly_service import AnomalyDetectionService

anomaly_bp = Blueprint('anomaly', __name__)
anomaly_service = AnomalyDetectionService()


@anomaly_bp.route('/detect', methods=['POST'])
def detect_anomaly():
    try:
        data = request.get_json() if request.is_json else {}
        
        result = anomaly_service.detect_anomaly(data)
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@anomaly_bp.route('/overtime', methods=['GET'])
def check_overtime():
    try:
        threshold_minutes = request.args.get('threshold', 1440, type=int)
        
        overtime_vehicles = anomaly_service.detect_overtime_parking(threshold_minutes)
        
        return jsonify({
            'success': True,
            'data': overtime_vehicles
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@anomaly_bp.route('/wrong_spot', methods=['POST'])
def check_wrong_spot():
    try:
        data = request.get_json() if request.is_json else {}
        
        result = anomaly_service.check_wrong_spot(
            data.get('license_plate'),
            data.get('current_spot_id'),
            data.get('reserved_spot_id')
        )
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@anomaly_bp.route('/suspicious', methods=['GET'])
def get_suspicious_vehicles():
    try:
        suspicious = anomaly_service.detect_suspicious_activity()
        
        return jsonify({
            'success': True,
            'data': suspicious
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@anomaly_bp.route('/logs', methods=['GET'])
def get_anomaly_logs():
    try:
        limit = request.args.get('limit', 100, type=int)
        anomaly_type = request.args.get('type', None)
        
        logs = anomaly_service.get_anomaly_logs(limit, anomaly_type)
        
        return jsonify({
            'success': True,
            'data': logs
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@anomaly_bp.route('/check_all', methods=['POST'])
def check_all_anomalies():
    try:
        data = request.get_json() if request.is_json else {}
        
        results = anomaly_service.check_all_anomalies(data)
        
        return jsonify({
            'success': True,
            'data': results
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500
