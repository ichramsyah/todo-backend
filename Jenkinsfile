pipeline {
  agent any
  environment {
    IMAGE = 'backend:latest'
    DEPLOY_DIR = '/opt/application'
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
        sh 'sleep 5 && docker exec backend wget -qO- http://127.0.0.1:8080/api/health'
      }
    }
  }
}
