node {
    try {
        stage('Checking out code') {
            checkout scm
        }
        def version = sh(
            script: 'make version',
            returnStdout: true
        ).trim()
        def props = [ version: version ]

        stage("Testing ${version}") {
            sh 'make test-docker'
        }
        stage("Build image ${version}") {
            sh 'make build-docker'
        }
        if (env.BRANCH_NAME == "master") {
            stage("Push image") {
                sh 'make push publish'
            }
        } else {
            def branchMatch = env.BRANCH_NAME =~ /^release-(0|[1-9]\d*)\.(0|[1-9]\d*)\.x$/
            if (branchMatch.find()) {
                def releaseVersion = readFile "${env.WORKSPACE}/operator-version"
                def versionMatch = releaseVersion =~ /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-(.+))?$+/
                if (!versionMatch.find()) {
                    error("Incorrect version ${releaseVersion} defined in ./operator-version")
                }
                if (versionMatch.group(1) != branchMatch.group(1) || versionMatch.group(2) != branchMatch.group(2)) {
                    error("Version ${releaseVersion} does not match branch it is being built on ${env.BRANCH_NAME}")
                }
                props.releaseVersion = releaseVersion
                stage("Publish Version ${releaseVersion}") {
                    sh "make push publishRelease RELEASE_VERSION=\"${releaseVersion}\""
                }
            }
        }
        writeFile file: 'build.properties', text: props.collect { k, v -> "${k}=${v}" }.join("\n")
        archiveArtifacts artifacts: 'build.properties'
    } catch (e) {
        slackSend color: 'danger', message: "Build of spinnaker-operator failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
        throw e
    }
}
