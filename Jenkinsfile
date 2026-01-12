pipeline {
  agent any

  environment {
    DOCKER_REPO  = "ksk5940/webk8s"
    DOCKER_CREDS = "dockerhub-creds"
    LATEST_TAG   = "latest"
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

    stage('Detect OS + Git Info') {
      steps {
        script {
          // OS detection
          env.IS_WINDOWS = isUnix() ? "false" : "true"

          // Get Git short commit
          if (isUnix()) {
            env.GIT_SHORT = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
          } else {
            env.GIT_SHORT = bat(script: "@git rev-parse --short HEAD", returnStdout: true).trim()
            env.GIT_SHORT = env.GIT_SHORT.replaceAll("\\s+", "")
          }

          // Image tags
          env.IMAGE_TAG = "${env.BUILD_NUMBER}"
          env.COMMIT_TAG = "${env.GIT_SHORT}"

          echo "OS Windows? ${env.IS_WINDOWS}"
          echo "Commit: ${env.GIT_SHORT}"
          echo "Tags: ${env.IMAGE_TAG}, ${env.COMMIT_TAG}, latest"
        }
      }
    }

    stage('Check Docker') {
      steps {
        script {
          if (isUnix()) {
            sh '''
              docker version
              docker info
            '''
          } else {
            bat '''
              docker version
              docker info
            '''
          }
        }
      }
    }

    stage('Build Docker Image') {
      steps {
        script {
          if (isUnix()) {
            sh """
              docker build ^
                -t ${DOCKER_REPO}:${IMAGE_TAG} ^
                -t ${DOCKER_REPO}:${LATEST_TAG} .
            """.replace("^", "\\")
          } else {
            bat """
              docker build ^
                -t %DOCKER_REPO%:%IMAGE_TAG% ^
                -t %DOCKER_REPO%:%LATEST_TAG% .
            """
          }
        }
      }
    }

    stage('Docker Login') {
      steps {
        withCredentials([usernamePassword(credentialsId: "${DOCKER_CREDS}",
                                          usernameVariable: 'DH_USER',
                                          passwordVariable: 'DH_PASS')]) {
          script {
            if (isUnix()) {
              sh '''
                echo "$DH_PASS" | docker login -u "$DH_USER" --password-stdin
              '''
            } else {
              bat '''
                echo %DH_PASS% | docker login -u %DH_USER% --password-stdin
              '''
            }
          }
        }
      }
    }

    stage('Push Image') {
      steps {
        script {
          if (isUnix()) {
            sh """
              docker push ${DOCKER_REPO}:${IMAGE_TAG}
              docker push ${DOCKER_REPO}:${LATEST_TAG}
            """
          } else {
            bat """
              docker push %DOCKER_REPO%:%IMAGE_TAG%
              docker push %DOCKER_REPO%:%LATEST_TAG%
            """
          }
        }
      }
    }
  }

  post {
    always {
      script {
        if (isUnix()) {
          sh '''
            docker logout || true
            docker image prune -af || true
          '''
        } else {
          bat '''
            docker logout || exit /b 0
            docker image prune -af || exit /b 0
          '''
        }
      }
    }
    success {
      echo "✅ Pushed: ${DOCKER_REPO}:${BUILD_NUMBER}, ${DOCKER_REPO}:${GIT_SHORT}, ${DOCKER_REPO}:latest"
    }
    failure {
      echo "❌ Build/Push failed"
    }
  }
}
