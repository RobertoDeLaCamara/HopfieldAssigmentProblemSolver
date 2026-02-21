pipeline {
    agent { label 'agent-45' }

    options {
        buildDiscarder(logRotator(numToKeepStr: '5'))
        timestamps()
        timeout(time: 60, unit: 'MINUTES')
    }

    environment {
        REGISTRY     = "192.168.1.86:5000"
        SOLVER_IMAGE = "hopfield-solver"
        API_IMAGE    = "hopfield-api"
        NO_PROXY     = 'localhost,127.0.0.1,192.168.1.0/24,192.168.1.86,192.168.1.62,192.168.1.45'
        no_proxy     = 'localhost,127.0.0.1,192.168.1.0/24,192.168.1.86,192.168.1.62,192.168.1.45'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Images') {
            steps {
                echo 'Building Docker images...'
                sh "docker build -t ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER} -t ${REGISTRY}/${SOLVER_IMAGE}:latest ./hopfield"
                sh "docker build -t ${REGISTRY}/${API_IMAGE}:${env.BUILD_NUMBER} -t ${REGISTRY}/${API_IMAGE}:latest ./api"
            }
        }

        stage('Code Quality Checks') {
            parallel {
                stage('Lint Python') {
                    steps {
                        sh """
                        docker run --rm --user root \
                            ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER} \
                            sh -c 'pip install --quiet flake8 && flake8 src/ --max-line-length=120 --count --statistics'
                        """
                    }
                }

                stage('Lint Go') {
                    steps {
                        sh """
                        docker run --rm \
                            golang:1.21-alpine \
                            sh -c 'apk add --no-cache git && cd /tmp && mkdir app && cd app && cp -r /dev/null . 2>/dev/null; true'
                        """
                        sh """
                        docker run --rm \
                            -v "${env.WORKSPACE}/api:/app" \
                            -w /app \
                            golang:1.21-alpine \
                            sh -c 'go vet ./...'
                        """
                    }
                }

                stage('Security Checks') {
                    steps {
                        catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                            echo 'Checking for security vulnerabilities...'
                            sh """
                            docker run --rm --user root \
                                ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER} \
                                sh -c 'pip install --quiet pip-audit && pip-audit -r requirements.txt'
                            """
                        }
                    }
                }
            }
        }

        stage('Run Tests') {
            parallel {
                stage('Python Tests') {
                    steps {
                        script {
                            try {
                                sh """
                                docker run --name test-solver-${env.BUILD_NUMBER} \
                                    --user root \
                                    ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER} \
                                    sh -c 'pip install --quiet pytest pytest-cov && python -m pytest tests/ -v \
                                        --junitxml=test-results-python.xml \
                                        --cov=src \
                                        --cov-report=xml:coverage-python.xml \
                                        --cov-report=term-missing \
                                        --disable-warnings'
                                """
                            } finally {
                                sh "docker cp test-solver-${env.BUILD_NUMBER}:/app/test-results-python.xml ${env.WORKSPACE}/test-results-python.xml || true"
                                sh "docker cp test-solver-${env.BUILD_NUMBER}:/app/coverage-python.xml ${env.WORKSPACE}/coverage-python.xml || true"
                                sh "docker rm test-solver-${env.BUILD_NUMBER} || true"
                            }
                        }
                    }
                    post {
                        always {
                            junit allowEmptyResults: true, testResults: 'test-results-python.xml'
                        }
                    }
                }

                stage('Go Tests') {
                    steps {
                        script {
                            try {
                                sh """
                                docker run --name test-api-${env.BUILD_NUMBER} \
                                    -v "${env.WORKSPACE}/api:/app" \
                                    -w /app \
                                    golang:1.21-alpine \
                                    sh -c 'apk add --no-cache git gcc musl-dev && go test ./... -v -coverprofile=coverage-go.out 2>&1 | tee test-output-go.txt; go tool cover -func=coverage-go.out'
                                """
                            } finally {
                                sh "docker cp test-api-${env.BUILD_NUMBER}:/app/coverage-go.out ${env.WORKSPACE}/coverage-go.out || true"
                                sh "docker rm test-api-${env.BUILD_NUMBER} || true"
                            }
                        }
                    }
                }
            }
        }

        stage('SonarQube Analysis') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'sonarqube-credentials',
                    usernameVariable: 'SONAR_USER',
                    passwordVariable: 'SONAR_PASS'
                )]) {
                    sh """
                        docker run --rm \
                            -e SONAR_USER="\$SONAR_USER" \
                            -e SONAR_PASS="\$SONAR_PASS" \
                            -v "${env.WORKSPACE}:/usr/src" \
                            sonarsource/sonar-scanner-cli \
                            -Dsonar.projectKey=HopfieldAssigmentProblemSolver \
                            -Dsonar.sources=hopfield/src,api \
                            -Dsonar.tests=hopfield/tests \
                            -Dsonar.python.version=3.11 \
                            -Dsonar.python.coverage.reportPaths=coverage-python.xml \
                            -Dsonar.host.url=http://192.168.1.86:9000 \
                            -Dsonar.login="\$SONAR_USER" \
                            -Dsonar.password="\$SONAR_PASS" \
                            -Dsonar.scm.disabled=true
                    """
                }
            }
        }

        stage('Push to Registry') {
            steps {
                echo "Pushing images to ${REGISTRY}..."
                sh "docker push ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER}"
                sh "docker push ${REGISTRY}/${SOLVER_IMAGE}:latest"
                sh "docker push ${REGISTRY}/${API_IMAGE}:${env.BUILD_NUMBER}"
                sh "docker push ${REGISTRY}/${API_IMAGE}:latest"
            }
        }
    }

    post {
        always {
            sh 'rm -f test-results-python.xml coverage-python.xml coverage-go.out || true'
            sh "docker rmi ${REGISTRY}/${SOLVER_IMAGE}:${env.BUILD_NUMBER} || true"
            sh "docker rmi ${REGISTRY}/${API_IMAGE}:${env.BUILD_NUMBER} || true"
            cleanWs()
        }
        success {
            echo 'Pipeline succeeded!'
        }
        failure {
            echo 'Pipeline failed.'
        }
    }
}
