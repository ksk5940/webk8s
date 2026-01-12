pipeline {
  agent any

  environment {
    DOCKER_REPO   = "ksk5940/webk8s"
    IMAGE_TAG     = "${BUILD_NUMBER}"
    LATEST_TAG    = "latest"
    DOCKER_CREDS  = "dockerhub-creds"
  }

  options {
    timestamps()
    disableConcurrentBuilds()
  }

  stages {

    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Build Docker Image') {
      steps {
        script {
          sh """
            docker version
            docker build -t ${DOCKER_REPO}:${IMAGE_TAG} -t ${DOCKER_REPO}:${LATEST_TAG} .
          """
        }
      }
    }

    stage('Docker Login') {
      steps {
        withCredentials([usernamePassword(credentialsId: "${DOCKER_CREDS}", usernameVariable: 'DH_USER', passwordVariable: 'DH_PASS')]) {
          sh """
            echo "${DH_PASS}" | docker login -u "${DH_USER}" --password-stdin
          """
        }
      }
    }

    stage('Push Image') {
      steps {
        sh """
          docker push ${DOCKER_REPO}:${IMAGE_TAG}
          docker push ${DOCKER_REPO}:${LATEST_TAG}
        """
      }
    }

  }

  post {
    always {
      sh "docker logout || true"
      sh "docker image prune -af || true"
    }
    success {
      echo "✅ Pushed: ${DOCKER_REPO}:${IMAGE_TAG} and ${DOCKER_REPO}:${LATEST_TAG}"
    }
    failure {
      echo "❌ Build/Push failed"
    }
  }
}
