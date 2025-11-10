

## Dev
* Docker
  * Start python server in Pycharm
  * connect to http://localhost:8000
* .venv
  * If running in .venv activate first `source .venv/bin/activate` 
  * Run this in the background `uvicorn main:app --port 8000 --reload`
  * connect to http://localhost:8000
* Frontend
  * Start Vite Dev in Pycharm
  * Open http://localhost:5173/

## Production
* API Server
  * If running in .venv activate first `source .venv/bin/activate` 
  * Run this in the background `uvicorn main:app --port 8000 --reload`
* Frontend
  * Use pm2 