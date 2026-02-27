from node:20.19.0-alpine

# Create app directory
WORKDIR /app/client-pwa

# Copy app source
COPY . /app

# Install dependencies
RUN npm install

# Add Poetry to PATH
ENV PATH="/app/client-pwa/node_modules/.bin:$PATH"

# Expose dev port
EXPOSE 5173

# Start app
CMD ["vite", "dev", "--host", "0.0.0.0"]