# build environment
FROM node:16.11.1-alpine as build
WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH
COPY package.json ./
COPY yarn.lock ./
COPY . ./
RUN npm run build

# production environment
FROM nginx:1.25.1-alpine
COPY --from=build /app/build /usr/share/nginx/html
# new
#COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
