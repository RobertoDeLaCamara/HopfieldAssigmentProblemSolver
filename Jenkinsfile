pipeline {
    agent any

    options {
        timestamps()
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
    }

    stages {
        stage('Prepare') {
            steps {
                sh 'mkdir -p reports'
            }
        }

        stage('Lint') {
            parallel {
                stage('Lint Python') {
                    steps {
                        sh '''
                            docker run --rm \
                                -v "${WORKSPACE}:/workspace" -w /workspace/hopfield \
                                python:3.11-slim sh -c '
                                    pip install --quiet --no-cache-dir flake8 black isort &&
                                    flake8 src/ --max-line-length=88 --extend-ignore=E203,W503 &&
                                    black --check src/ &&
                                    isort --profile black --check-only src/
                                '
                        '''
                    }
                }
                stage('Lint Go') {
                    steps {
                        sh '''
                            docker run --rm \
                                -v "${WORKSPACE}:/workspace" -w /workspace/api \
                                golangci/golangci-lint:v1.55.2 \
                                golangci-lint run ./...
                        '''
                    }
                }
            }
        }

        stage('Test') {
            parallel {
                stage('Test Python') {
                    steps {
                        sh '''
                            docker run --rm \
                                -v "${WORKSPACE}:/workspace" -w /workspace \
                                python:3.11-slim sh -c '
                                    cd hopfield &&
                                    pip install --quiet --no-cache-dir \
                                        -r requirements.txt \
                                        -r requirements-test.txt &&
                                    python -m pytest tests/ -v \
                                        --junitxml=../reports/python-junit.xml \
                                        --cov=src \
                                        --cov-report=xml:../reports/python-coverage.xml \
                                        --cov-report=term
                                '
                        '''
                    }
                    post {
                        always {
                            junit allowEmptyResults: true, testResults: 'reports/python-junit.xml'
                        }
                    }
                }
                stage('Test Go') {
                    steps {
                        sh '''
                            docker run --rm \
                                -v "${WORKSPACE}:/workspace" -w /workspace \
                                golang:1.21-alpine sh -c '
                                    mkdir -p reports &&
                                    cd api &&
                                    go mod download &&
                                    go test -v -race \
                                        -coverprofile=../reports/go-coverage.txt \
                                        -covermode=atomic ./... 2>&1 | tee ../reports/go-test.txt &&
                                    go install github.com/jstemmer/go-junit-report/v2@latest &&
                                    cat ../reports/go-test.txt | $(go env GOPATH)/bin/go-junit-report > ../reports/go-junit.xml
                                '
                        '''
                    }
                    post {
                        always {
                            junit allowEmptyResults: true, testResults: 'reports/go-junit.xml'
                        }
                    }
                }
            }
        }

        stage('Build Images') {
            steps {
                sh '''
                    docker build -t hopfield-api-gateway:${BUILD_NUMBER} \
                                 -t hopfield-api-gateway:latest ./api
                    docker build -t hopfield-service:${BUILD_NUMBER} \
                                 -t hopfield-service:latest ./hopfield
                '''
            }
        }

        stage('Integration Tests') {
            steps {
                sh '''
                    docker compose up -d --build hopfield-service api-gateway

                    echo "Waiting for services to be healthy..."
                    timeout 120 sh -c '
                        until [ "$(docker inspect --format={{.State.Health.Status}} hopfield-assignment-solver 2>/dev/null)" = "healthy" ]; do
                            sleep 3
                        done
                    '
                    echo "Hopfield Service: healthy"

                    timeout 120 sh -c '
                        until [ "$(docker inspect --format={{.State.Health.Status}} hopfield-api-gateway 2>/dev/null)" = "healthy" ]; do
                            sleep 3
                        done
                    '
                    echo "API Gateway: healthy"
                '''
                sh '''
                    docker run --rm --network host \
                        -v "${WORKSPACE}/tests:/tests" \
                        -v "${WORKSPACE}/reports:/reports" \
                        python:3.11-slim sh -c '
                            pip install --quiet --no-cache-dir pytest requests &&
                            python -m pytest /tests/integration_test.py -v \
                                --junitxml=/reports/integration-junit.xml
                        '
                '''
            }
            post {
                always {
                    sh 'docker compose down -v || true'
                    junit allowEmptyResults: true, testResults: 'reports/integration-junit.xml'
                }
                failure {
                    sh 'docker compose logs --tail=100 || true'
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true
        }
        success {
            echo 'Pipeline completed successfully!'
        }
        failure {
            echo 'Pipeline failed. Check stage logs for details.'
        }
    }
}
