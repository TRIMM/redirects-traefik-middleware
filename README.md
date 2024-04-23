# TRIMM Redirects Traefik Middleware

## Introduction
This project is a traefik middleware written in GoLang.
The purpose of it is to redirect according to records processed from the Central API backend.

## Instructions

### 1. Create an .env file
In the project directory you need to create a `.env` file. You can copy the values from
the `.env.example` file and overwrite the values with valid ones.

### 2. Make sure the Central API backend is running
For this, there is an instruction manual in the following project:
https://gitlab.trimm.nl/technology/platform-applications/redneck

### 3. Run the application
If you are using GoLand, then you can just run the application using the play or debug button.

In any other case, open the terminal in the current directory, and you can just run the `go run .` command.