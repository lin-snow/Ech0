FROM public.ecr.aws/docker/library/node:20-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install --production
COPY server.js ./
ENV NOTE_HOST=http://localhost:8080
CMD ["node","server.js"]
