@NonCPS
def getReleaseVersion(branchName) {
    def branchMatch = env.BRANCH_NAME =~ /^release-(0|[1-9]\d*)\.(0|[1-9]\d*)\.x$/
    if (branchMatch.find()) {
        def releaseVersion = readFile("operator-version").split("\\n")[0].trim()
        print "Found committed version file contents: ${releaseVersion}"

        def versionMatch = releaseVersion =~ /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-(.+))?+$/
        if (!versionMatch.find()) {
            error("Incorrect version ${releaseVersion} defined in ./operator-version")
        }
        if (versionMatch.group(1) != branchMatch.group(1) || versionMatch.group(2) != branchMatch.group(2)) {
            error("Version ${releaseVersion} does not match branch it is being built on ${env.BRANCH_NAME}")
        }
        return branchMatch.group(0)
    }
}

node {
    try {
        stage('Checking out code') {
            checkout scm
        }
        def version = sh(
            script: 'make version',
            returnStdout: true
        ).trim()
        def props = [ version: version, buildArgs: "--" ]

        def releaseVersion = getReleaseVersion(env.BRANCH_NAME)

        if(releaseVersion){
           props.releaseVersion = releaseVersion
           props.buildArgs="RELEASE_VERSION=\"${releaseVersion}\""
        }

        stage("Testing ${version}") {
            sh 'make test-docker'
        }
        stage("Build image ${version}") {
            sh "make build-docker ${buildArgs}"
        }
        if (env.BRANCH_NAME == "master") {
            stage("Push image") {
                sh 'make push publish'
            }
        } else {
            if (releaseVersion) {
                stage("Publish Version ${releaseVersion}") {
                    sh "make push publishRelease ${buildArgs}"
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
