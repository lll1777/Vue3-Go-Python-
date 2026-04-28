import random
from datetime import datetime, timedelta
from typing import Dict, List, Optional


class AnomalyDetectionService:
    def __init__(self):
        self.anomaly_logs: List[Dict] = []
        self.overtime_threshold_minutes = 1440
        self.wrong_spot_confidence_threshold = 0.8
    
    def detect_anomaly(self, data: Dict) -> Dict:
        anomaly_type = data.get('type', 'unknown')
        license_plate = data.get('license_plate')
        spot_id = data.get('spot_id')
        
        anomalies = []
        
        if 'parking_duration' in data:
            overtime_anomaly = self._check_overtime(data['parking_duration'], license_plate, spot_id)
            if overtime_anomaly:
                anomalies.append(overtime_anomaly)
        
        if 'expected_spot' in data and 'actual_spot' in data:
            wrong_spot_anomaly = self._check_wrong_spot(
                data['actual_spot'],
                data['expected_spot'],
                license_plate
            )
            if wrong_spot_anomaly:
                anomalies.append(wrong_spot_anomaly)
        
        if 'entry_count' in data:
            suspicious_anomaly = self._check_suspicious_activity(
                data['entry_count'],
                license_plate,
                data.get('time_window_hours', 24)
            )
            if suspicious_anomaly:
                anomalies.append(suspicious_anomaly)
        
        result = {
            'has_anomaly': len(anomalies) > 0,
            'anomalies': anomalies,
            'detected_at': datetime.now().isoformat()
        }
        
        if len(anomalies) > 0:
            self._log_anomaly(result)
        
        return result
    
    def _check_overtime(self, duration_minutes: int, license_plate: str = None, spot_id: str = None) -> Optional[Dict]:
        if duration_minutes > self.overtime_threshold_minutes:
            return {
                'type': 'overtime',
                'severity': 'high' if duration_minutes > self.overtime_threshold_minutes * 2 else 'medium',
                'license_plate': license_plate,
                'spot_id': spot_id,
                'duration_minutes': duration_minutes,
                'threshold_minutes': self.overtime_threshold_minutes,
                'message': f'车辆停车时长 {duration_minutes} 分钟，超过阈值 {self.overtime_threshold_minutes} 分钟'
            }
        return None
    
    def _check_wrong_spot(self, actual_spot: str, expected_spot: str, license_plate: str = None) -> Optional[Dict]:
        if actual_spot != expected_spot:
            return {
                'type': 'wrong_spot',
                'severity': 'high',
                'license_plate': license_plate,
                'actual_spot': actual_spot,
                'expected_spot': expected_spot,
                'confidence': 1.0,
                'message': f'车辆 {license_plate} 停在错误的车位：实际 {actual_spot}，预期 {expected_spot}'
            }
        return None
    
    def _check_suspicious_activity(self, entry_count: int, license_plate: str = None, time_window_hours: int = 24) -> Optional[Dict]:
        if entry_count > 5:
            return {
                'type': 'suspicious_activity',
                'severity': 'medium' if entry_count < 10 else 'high',
                'license_plate': license_plate,
                'entry_count': entry_count,
                'time_window_hours': time_window_hours,
                'threshold': 5,
                'message': f'车辆 {license_plate} 在 {time_window_hours} 小时内进出 {entry_count} 次，活动异常'
            }
        return None
    
    def detect_overtime_parking(self, threshold_minutes: int = 1440) -> List[Dict]:
        simulated_overtime = [
            {
                'id': f'overtime-{i+1}',
                'license_plate': f'京{chr(65+i)}{random.randint(10000, 99999)}',
                'spot_number': f'A{random.randint(1, 50)}',
                'parking_duration': threshold_minutes + random.randint(60, 1440),
                'entry_time': (datetime.now() - timedelta(minutes=threshold_minutes + random.randint(60, 1440))).isoformat(),
                'severity': 'high' if random.random() > 0.5 else 'medium',
                'detected_at': datetime.now().isoformat()
            }
            for i in range(random.randint(0, 3))
        ]
        
        return simulated_overtime
    
    def check_wrong_spot(self, license_plate: str, current_spot_id: str, reserved_spot_id: str = None) -> Dict:
        if not reserved_spot_id:
            return {
                'is_anomaly': False,
                'message': 'No reserved spot to compare'
            }
        
        is_wrong_spot = current_spot_id != reserved_spot_id
        
        result = {
            'is_anomaly': is_wrong_spot,
            'license_plate': license_plate,
            'current_spot': current_spot_id,
            'reserved_spot': reserved_spot_id,
            'type': 'wrong_spot' if is_wrong_spot else 'normal',
            'severity': 'high' if is_wrong_spot else 'none',
            'message': f'车辆 {license_plate} 停在 {current_spot_id}，预约车位是 {reserved_spot_id}' if is_wrong_spot else '车辆停在正确的车位',
            'detected_at': datetime.now().isoformat()
        }
        
        if is_wrong_spot:
            self._log_anomaly({'has_anomaly': True, 'anomalies': [result]})
        
        return result
    
    def detect_suspicious_activity(self) -> List[Dict]:
        simulated_suspicious = []
        
        if random.random() > 0.7:
            simulated_suspicious.append({
                'id': f'suspicious-1',
                'license_plate': f'京{chr(65+random.randint(0, 25))}{random.randint(10000, 99999)}',
                'activity_type': 'frequent_entry_exit',
                'entry_count': random.randint(6, 15),
                'time_window_hours': 24,
                'severity': 'medium',
                'detected_at': datetime.now().isoformat()
            })
        
        return simulated_suspicious
    
    def check_all_anomalies(self, data: Dict) -> Dict:
        all_anomalies = []
        
        overtime = self.detect_overtime_parking()
        all_anomalies.extend([{**a, 'type': 'overtime'} for a in overtime])
        
        suspicious = self.detect_suspicious_activity()
        all_anomalies.extend([{**a, 'type': 'suspicious_activity'} for a in suspicious])
        
        if 'wrong_spot_checks' in data:
            for check in data['wrong_spot_checks']:
                result = self.check_wrong_spot(
                    check.get('license_plate'),
                    check.get('current_spot_id'),
                    check.get('reserved_spot_id')
                )
                if result.get('is_anomaly'):
                    all_anomalies.append(result)
        
        high_count = sum(1 for a in all_anomalies if a.get('severity') == 'high')
        medium_count = sum(1 for a in all_anomalies if a.get('severity') == 'medium')
        
        return {
            'has_anomalies': len(all_anomalies) > 0,
            'total_anomalies': len(all_anomalies),
            'high_severity': high_count,
            'medium_severity': medium_count,
            'anomalies': all_anomalies,
            'checked_at': datetime.now().isoformat()
        }
    
    def _log_anomaly(self, result: Dict):
        log_entry = {
            **result,
            'timestamp': datetime.now().isoformat()
        }
        self.anomaly_logs.insert(0, log_entry)
        
        if len(self.anomaly_logs) > 500:
            self.anomaly_logs = self.anomaly_logs[:500]
    
    def get_anomaly_logs(self, limit: int = 100, anomaly_type: str = None) -> List[Dict]:
        filtered_logs = self.anomaly_logs
        
        if anomaly_type:
            filtered_logs = [
                log for log in filtered_logs
                if any(a.get('type') == anomaly_type for a in log.get('anomalies', []))
            ]
        
        return filtered_logs[:limit]
