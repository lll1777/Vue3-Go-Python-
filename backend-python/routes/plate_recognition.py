from flask import Blueprint, request, jsonify
from services.plate_service import PlateRecognitionService

plate_bp = Blueprint('plate', __name__)
plate_service = PlateRecognitionService()


@plate_bp.route('/recognize', methods=['POST'])
def recognize_plate():
    try:
        if 'image' in request.files:
            image_file = request.files['image']
            result = plate_service.recognize_from_image(image_file)
        elif request.is_json:
            data = request.get_json()
            if 'image_base64' in data:
                result = plate_service.recognize_from_base64(data['image_base64'])
            elif 'license_plate' in data:
                result = plate_service.recognize_from_text(data['license_plate'])
            else:
                return jsonify({
                    'success': False,
                    'message': 'Invalid request: need image, image_base64, or license_plate'
                }), 400
        else:
            return jsonify({
                'success': False,
                'message': 'Invalid request format'
            }), 400
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@plate_bp.route('/verify', methods=['POST'])
def verify_plate():
    try:
        data = request.get_json() if request.is_json else {}
        
        license_plate = data.get('license_plate')
        expected_plate = data.get('expected_plate')
        
        if not license_plate or not expected_plate:
            return jsonify({
                'success': False,
                'message': 'Missing license_plate or expected_plate'
            }), 400
        
        result = plate_service.verify_plate(license_plate, expected_plate)
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@plate_bp.route('/logs', methods=['GET'])
def get_plate_logs():
    try:
        limit = request.args.get('limit', 100, type=int)
        logs = plate_service.get_recognition_logs(limit)
        
        return jsonify({
            'success': True,
            'data': logs
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500


@plate_bp.route('/validate', methods=['POST'])
def validate_plate():
    try:
        data = request.get_json() if request.is_json else {}
        
        license_plate = data.get('license_plate')
        
        if not license_plate:
            return jsonify({
                'success': False,
                'message': 'Missing license_plate'
            }), 400
        
        result = plate_service.validate_plate_format(license_plate)
        
        return jsonify({
            'success': True,
            'data': result
        })
    except Exception as e:
        return jsonify({
            'success': False,
            'message': str(e)
        }), 500
