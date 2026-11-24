pipeline {
  agent any
  environment {
    IMAGE = 'todo-backend:latest'
    DEPLOY_DIR = '/opt/todo-app'
  }
  stages {
    stage('Checkout') {
      steps { checkout scm }
    }
    stage('Build image') {
      steps { sh 'docker build -t $IMAGE .' }
    }
    stage('Deploy') {
      steps {
        sh 'docker compose -f $DEPLOY_DIR/docker-compose.yml up -d --no-deps --force-recreate backend'
      }
    }
    stage('Smoke test') {
      steps {
        sh 'sleep 5 && docker exec todo-backend wget -qO- http://localhost:8080/api/health'
      }
    }
  }
}
