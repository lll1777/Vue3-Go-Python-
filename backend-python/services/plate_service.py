import re
import random
from datetime import datetime
from typing import Dict, List, Optional


class PlateRecognitionService:
    def __init__(self):
        self.recognition_logs: List[Dict] = []
        self.province_codes = {
            '京': '北京', '津': '天津', '沪': '上海', '渝': '重庆',
            '冀': '河北', '晋': '山西', '辽': '辽宁', '吉': '吉林',
            '黑': '黑龙江', '苏': '江苏', '浙': '浙江', '皖': '安徽',
            '闽': '福建', '赣': '江西', '鲁': '山东', '豫': '河南',
            '鄂': '湖北', '湘': '湖南', '粤': '广东', '桂': '广西',
            '琼': '海南', '川': '四川', '贵': '贵州', '云': '云南',
            '藏': '西藏', '陕': '陕西', '甘': '甘肃', '青': '青海',
            '宁': '宁夏', '新': '新疆', '蒙': '内蒙古'
        }
        
        self.plate_patterns = {
            'standard': r'^[\u4e00-\u9fa5]{1}[A-Z]{1}[A-Z0-9]{5}$',
            'new_energy': r'^[\u4e00-\u9fa5]{1}[A-Z]{1}[A-Z0-9]{6}$',
            'military': r'^[A-Z]{1}[0-9]{6}$',
            'police': r'^[\u4e00-\u9fa5]{1}[A-Z]{1}[0-9]{4}警$'
        }
    
    def recognize_from_image(self, image_file) -> Dict:
        simulated_plates = [
            '京A12345', '京B67890', '沪C11111', '粤D22222',
            '浙E33333', '苏F44444', '川G55555', '鲁H66666'
        ]
        
        plate = random.choice(simulated_plates)
        confidence = round(0.9 + random.random() * 0.099, 3)
        
        result = {
            'license_plate': plate,
            'confidence': confidence,
            'plate_type': self._detect_plate_type(plate),
            'province': self._get_province(plate),
            'city_code': plate[1] if len(plate) > 1 else '',
            'recognized_at': datetime.now().isoformat()
        }
        
        self._log_recognition(result)
        
        return result
    
    def recognize_from_base64(self, image_base64: str) -> Dict:
        return self.recognize_from_image(None)
    
    def recognize_from_text(self, text: str) -> Dict:
        if not text:
            raise ValueError('Text cannot be empty')
        
        cleaned_text = re.sub(r'[\s\-_\.\(\)\[\]]', '', text.upper())
        
        is_valid, plate_type = self.validate_plate_format(cleaned_text)
        
        if not is_valid:
            return {
                'license_plate': cleaned_text,
                'confidence': 0.5,
                'plate_type': 'unknown',
                'province': None,
                'city_code': None,
                'is_valid': False,
                'recognized_at': datetime.now().isoformat()
            }
        
        result = {
            'license_plate': cleaned_text,
            'confidence': 1.0,
            'plate_type': plate_type,
            'province': self._get_province(cleaned_text),
            'city_code': cleaned_text[1] if len(cleaned_text) > 1 else '',
            'is_valid': True,
            'recognized_at': datetime.now().isoformat()
        }
        
        self._log_recognition(result)
        
        return result
    
    def _detect_plate_type(self, plate: str) -> str:
        if re.match(self.plate_patterns['new_energy'], plate):
            return 'new_energy'
        elif re.match(self.plate_patterns['police'], plate):
            return 'police'
        elif re.match(self.plate_patterns['military'], plate):
            return 'military'
        elif re.match(self.plate_patterns['standard'], plate):
            return 'standard'
        return 'unknown'
    
    def _get_province(self, plate: str) -> Optional[str]:
        if not plate:
            return None
        
        first_char = plate[0]
        
        if first_char in self.province_codes:
            return self.province_codes[first_char]
        
        return None
    
    def validate_plate_format(self, plate: str) -> tuple:
        if not plate:
            return False, 'empty'
        
        for plate_type, pattern in self.plate_patterns.items():
            if re.match(pattern, plate):
                return True, plate_type
        
        return False, 'invalid'
    
    def verify_plate(self, license_plate: str, expected_plate: str) -> Dict:
        if not license_plate or not expected_plate:
            return {
                'match': False,
                'similarity': 0.0,
                'message': 'Missing license plate or expected plate'
            }
        
        license_plate_clean = re.sub(r'[\s\-_]', '', license_plate.upper())
        expected_plate_clean = re.sub(r'[\s\-_]', '', expected_plate.upper())
        
        match = license_plate_clean == expected_plate_clean
        
        similarity = self._calculate_similarity(license_plate_clean, expected_plate_clean)
        
        return {
            'match': match,
            'similarity': similarity,
            'recognized_plate': license_plate_clean,
            'expected_plate': expected_plate_clean,
            'verified_at': datetime.now().isoformat()
        }
    
    def _calculate_similarity(self, str1: str, str2: str) -> float:
        if len(str1) != len(str2):
            return 0.0
        
        matches = sum(1 for a, b in zip(str1, str2) if a == b)
        return matches / len(str1)
    
    def _log_recognition(self, result: Dict):
        log_entry = {
            **result,
            'timestamp': datetime.now().isoformat()
        }
        self.recognition_logs.insert(0, log_entry)
        
        if len(self.recognition_logs) > 1000:
            self.recognition_logs = self.recognition_logs[:1000]
    
    def get_recognition_logs(self, limit: int = 100) -> List[Dict]:
        return self.recognition_logs[:limit]
