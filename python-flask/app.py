import os
from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/')
def hello():
    return jsonify({
        "status": "success",
        "message": "Halo dari Python Flask di dalam Docker!",
        "framework": "Flask",
        "debug_mode": os.environ.get("FLASK_DEBUG", "False")
    })

if __name__ == '__main__':
    # Mengambil port dari environment variable, default 5000
    port = int(os.environ.get("PORT", 5000))
    # Host 0.0.0.0 sangat penting agar container bisa diakses dari luar
    app.run(host='0.0.0.0', port=port)
