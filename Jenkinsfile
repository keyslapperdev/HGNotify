@Library('eigi-jenkins-library@master')_

def pipeline = new hostgator.v1.Pipeline(this)
def os_util = new common.v1.OsUtil(this)

def configMap = [appName: 'hgnotify',
                    gchatNotifications: [
                      room: 'https://chat.googleapis.com/v1/spaces/AAAARQ66BnQ/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=3hUfDafi6WcUFT6R3szJSk3m3sMwOI6sxi8VyAbrVd4%3D&threadKey=jenkins'
                    ]
                ]

pipeline.hg(configMap){
    def testPod = null

    stage('Interrupt any running builds') {
        os_util.stopRunningPipeline()
    }
    stage('Build Image') {
        openshift.withCluster(){
            openshift.withProject(){
                openshift.raw( 'rollout latest dc/hgnotify-dev')
             }
        }
        os_util.buildImage( 5 )
    }
    stage('Wait for Deployment') {
        testPod = os_util.waitForDeploymentRollout()
    }
    stage('Check Tests') {
        openshift.withCluster() {
            openshift.withProject() {
                def metaName = testPod.object().metadata.name

                testPod.exec("${metaName} -c ${env.APP_NAME} -- bash -c 'export RUN_INTEGRATION=true; export GOPATH=/tmp; cd /tmp/src/; go test ./... -v > ../test_results || echo Tests returned exit code: \$?'")
                testPod.exec("${metaName} -c ${env.APP_NAME} -- bash -c 'cd /tmp/;/tmp/src/bin/go2xunit -input test_results -output test_results.xml'")
                openshift.raw("rsync ${metaName}:/tmp/test_results.xml ./")
                junit "test_results.xml"
                if(currentBuild.currentResult != "SUCCESS"){
                    error("Tests Failed")
                }
            }
        }
    }
    stage('Tag Image for Development') {
        os_util.tagImage( 'development' );
    }
    stage('Wait for Approval to Push to Production') {
        os_util.requestApproval( 'production' );
    }
    stage('Tag Image for Production') {
        os_util.tagImage( 'production' );
    }
}
