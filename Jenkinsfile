node {
    try {
        stage('Checking out code') {
            checkout scm
        }
        def version = sh(
            script: 'make version',
            returnStdout: true
        ).trim()

        stage("Testing ${version}") {
            sh 'make test-docker'
        }
        stage("Build image ${version}") {
            sh 'make build-docker'
        }
        def branch = sh(
            script: 'git symbolic-ref --short HEAD || echo pr-branch',
            returnStdout: true
        ).trim()

        if (branch == 'master') {
            stage("Push image") {
                sh 'make push publish'
            }
        }
        def props = [ version: version ]
        writeFile file: 'build.properties', text: props.collect { k, v -> "${k}=${v}" }.join("\n")
        archiveArtifacts artifacts: 'build.properties'
    } catch (e) {
        slackSend color: 'danger', message: "Build of spinnaker-operator failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
        throw e
    }
}
