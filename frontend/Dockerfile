FROM node:22.12-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
RUN apk add --no-cache git python3 build-base || true
RUN npm ci --silent || npm install --no-audit --no-fund --legacy-peer-deps
COPY . .
ENV HOST=0.0.0.0
ENV VITE_HOST=0.0.0.0
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
