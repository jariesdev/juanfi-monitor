

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
  * Build, `cd client-pwa && ./node_modules/.bin/vite build`
    * Use pm2 
      * `PORT=8001 API_URL=http://localhost:8001 pm2 start build/index.js --interpreter node --watch --name pwifi`
    * PORT: Backend API port
    * API_URL: API URL to be use by Frontend