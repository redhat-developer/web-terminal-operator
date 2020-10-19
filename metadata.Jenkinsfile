#!/usr/bin/env groovy

def SOURCE_REPO = "redhat-developer/web-terminal-operator" //source repo from which to find and sync commits to pkgs.devel repo
def DWNSTM_REPO = "containers/web-terminal-dev-operator-metadata" // dist-git repo to use as target for everything

def SOURCE_BRANCH = "v1.0.x"         // target branch in GH repo, eg., master or v1.0.x
def DWNSTM_BRANCH = "web-terminal-1.0-rhel-8" // target branch in dist-git repo, eg., web-terminal-1.0-rhel-8

def buildNode = "rhel7-releng" // slave label
timeout(120) {
  node("${buildNode}"){
    stage "Sync repos"
    wrap([$class: 'TimestamperBuildWrapper']) {
      cleanWs()
      withCredentials([string(credentialsId:'devstudio-release.token', variable: 'GITHUB_TOKEN'), 
      file(credentialsId: 'crw-build.keytab', variable: 'CRW_KEYTAB')]) {
        checkout([$class: 'GitSCM',
          branches: [[name: "${SOURCE_BRANCH}"]],
          doGenerateSubmoduleConfigurations: false,
          credentialsId: 'devstudio-release',
          poll: true,
          extensions: [
            [$class: 'RelativeTargetDirectory', relativeTargetDir: "sources"],
          ],
          submoduleCfg: [],
          userRemoteConfigs: [[url: "https://github.com/${SOURCE_REPO}.git"]]])

          def BOOTSTRAP = '''
          export GITHUB_TOKEN=''' + GITHUB_TOKEN + '''
          export SOURCE_REPO=''' + SOURCE_REPO + '''
          export SOURCE_BRANCH=''' + SOURCE_BRANCH + '''
          export DWNSTM_REPO=''' + DWNSTM_REPO + '''
          export DWNSTM_BRANCH=''' + DWNSTM_BRANCH + '''
          export CRW_KEYTAB=''' + CRW_KEYTAB + '''
          export USER=''' + USER + '''
          curl -L -s -S https://raw.githubusercontent.com/redhat-developer/web-terminal-operator/v1.0.x/bootstrap-sync.sh -o ./sync.sh
          chmod +x ./sync.sh
          ./sync.sh
          '''

          sh BOOTSTRAP + '''
          # initialize kerberos
          export KRB5CCNAME=/var/tmp/crw-build_ccache
          kinit "crw-build/codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com@REDHAT.COM" -kt ''' + CRW_KEYTAB + '''
          klist # verify working

          cd ${WORKSPACE}/sources
          SOURCE_SHA=$(git rev-parse HEAD)
          cd ${WORKSPACE}/targetdwn

          # Update the base image
          curl -L -s -S https://raw.githubusercontent.com/redhat-developer/codeready-workspaces/master/product/updateBaseImages.sh -o /tmp/updateBaseImages.sh
          chmod +x /tmp/updateBaseImages.sh
          /tmp/updateBaseImages.sh -b ''' + DWNSTM_BRANCH + ''' --nocommit

          # Change all references of quay.io/wto to registry-proxy.engineering.redhat.com/rh-osbs
          sed -i -e 's/quay.io/registry-proxy.engineering.redhat.com/g' \
                  -e 's/wto/rh-osbs/g' \
                  manifests/web-terminal.clusterserviceversion.yaml
          if [[ $(git diff --name-only) ]]; then # file changed
            git add .
            git commit -s -m "[sync] Updated from ${SOURCE_REPO} @ ${SOURCE_SHA:0:8} " || true
            git push origin ${DWNSTM_BRANCH} || true
            echo "[sync] Updated from ${SOURCE_REPO} @ ${SOURCE_SHA:0:8} "
          else
            echo "Source and downstream contents are the same. No need to sync"
          fi
          '''
      }
    }
  }
}
