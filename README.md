# Prompt

## Context
I need to test a Cloud Run feature that allows deployment to Cloud Run using a Docker Compose file.
I would therefore like to create a typical application that I use quite frequently.

The application we want to set up is a list management tool for what has been eaten during a meal.
The information is stored in a Postgresql database made available by a Rest API with a Redis cache.

## To do
I want to create two applications: a front-end called “fe” and a back-end called “be.”
The “fe” application is in the “fe” folder and uses the JS language and the AngularJS framework.
The “be” application is in the “be” folder and uses the Go language.

The “be” application also connects to a Redis server and a Postgresql database server.
This application provides a Rest API.

## Requirements
This entire stack must work using a docker-compse.yml file and the docker compose up command.
