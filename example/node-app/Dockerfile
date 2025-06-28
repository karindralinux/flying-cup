FROM node:18-alpine

WORKDIR /app

COPY package.json .

RUN npm install

COPY . .

# TODO: forward request and respons to port 3000 in container to port 3000 in host
EXPOSE 3000 

CMD ["node", "server.js"]