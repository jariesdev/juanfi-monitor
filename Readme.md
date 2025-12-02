

## Dev
* Docker
  * Start python server in Pycharm
  * connect to http://localhost:8000
* .venv
  * If running in .venv activate first `source .venv/bin/activate` 
  * Run this in the background `uvicorn main:app --port 8000 --reload`
  * connect to http://localhost:8000
* Database Migration
  * Create models in sql_app/models
  * Auto generate migration file `alembic revision --autogenerate -m "Descriptive message for your migration"`
  * Update the file if needed
  * Run migration `alembic upgrade head`
  * to rollback, `alembic downgrade <target_revision_id>`
* Frontend
  * Start Vite Dev in Pycharm
  * Open http://localhost:5173/

## Production
* API Server
  * Create Project Directory and Virtual Environment `python3 -m venv .venv` (`sudo apt install python3-venv`)
  * If running in .venv activate first `source .venv/bin/activate`
  * Run migration, `alembic upgrade head`
  * Single instance: Run this in the background `uvicorn main:app --port 8000 --reload`
  * Multiple instance: `gunicorn -w 2 -k uvicorn.workers.UvicornWorker main:app -b 127.0.0.1:8000 --reload`
    * /etc/systemd/system/gunicorn.service
      ```
        [Unit]
        Description=Gunicorn daemon for VendorReport
        After=network.target
    
        [Service]
        User=root
        Group=root
        WorkingDirectory=/opt/vendoreport
        ExecStart=/project/path/.venv/bin/gunicorn -w 2 -k uvicorn.workers.UvicornWorker main:app --bind 127.0.0.1:8000 --reload
        EnvironmentFile=/opt/vendoreport/.env
    
        [Install]
        WantedBy=multi-user.target
        ```

* Frontend
  * Build, `cd client-pwa && ./node_modules/.bin/vite build`
    * Use pm2 
      * `PORT=8001 API_URL=http://localhost:8001 pm2 start build/index.js --interpreter node --watch --name pwifi`
    * PORT: Backend API port
    * API_URL: API URL to be use by Frontend