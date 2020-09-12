#!/usr/bin/env groovy

def TOOLING_REPO = "containers/web-terminal-tooling" // dist-git repo to use as target for everything
def EXEC_REPO = "containers/web-terminal-exec" // dist-git repo to use as target for everything
def OPERATOR_REPO = "containers/web-terminal" // dist-git repo to use as target for everything
def OPERATOR_METADATA_REPO = "containers/web-terminal-dev-operator-metadata" // dist-git repo to use as target for everything
def SOURCE_BRANCH = "web-terminal-1.0-rhel-8"

def buildNode = "rhel7-releng" // slave label
timeout(120) {
  node("${buildNode}"){
    stage "Sync repos"
    wrap([$class: 'TimestamperBuildWrapper']) {
      cleanWs()
      withCredentials([string(credentialsId:'devstudio-release.token', variable: 'GITHUB_TOKEN'), 
      file(credentialsId: 'crw-build.keytab', variable: 'CRW_KEYTAB')]) {

        // Login to quay robot
        withCredentials([usernamePassword(credentialsId: 'web_terminal_quay_robot', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
          // Install all pre-reqs
          sh 'sudo yum install podman -y'
          sh 'sudo podman login quay.io -u ${USERNAME} -p ${PASSWORD}'

          def BOOTSTRAP = '''
          # bootstrapping: if keytab is lost, upload to
          # https://codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com/credentials/store/system/domain/_/
          # then set Use secret text above and set Bindings > Variable (path to the file) as ''' + CRW_KEYTAB + '''
          chmod 700 ''' + CRW_KEYTAB + ''' && chown ''' + USER + ''' ''' + CRW_KEYTAB + '''
          # create .k5login file
          echo "crw-build/codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com@REDHAT.COM" > ~/.k5login
          chmod 644 ~/.k5login && chown ''' + USER + ''' ~/.k5login
          echo "pkgs.devel.redhat.com,10.16.101.66 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAplqWKs26qsoaTxvWn3DFcdbiBxqRLhFngGiMYhbudnAj4li9/VwAJqLm1M6YfjOoJrj9dlmuXhNzkSzvyoQODaRgsjCG5FaRjuN8CSM/y+glgCYsWX1HFZSnAasLDuW0ifNLPR2RBkmWx61QKq+TxFDjASBbBywtupJcCsA5ktkjLILS+1eWndPJeSUJiOtzhoN8KIigkYveHSetnxauxv1abqwQTk5PmxRgRt20kZEFSRqZOJUlcl85sZYzNC/G7mneptJtHlcNrPgImuOdus5CW+7W49Z/1xqqWI/iRjwipgEMGusPMlSzdxDX4JzIx6R53pDpAwSAQVGDz4F9eQ==
          " >> ~/.ssh/known_hosts
          ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
          # see https://mojo.redhat.com/docs/DOC-1071739
          if [[ -f ~/.ssh/config ]]; then mv -f ~/.ssh/config{,.BAK}; fi
          echo "
          GSSAPIAuthentication yes
          GSSAPIDelegateCredentials yes
          Host pkgs.devel.redhat.com
          User crw-build/codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com@REDHAT.COM
          " > ~/.ssh/config
          chmod 600 ~/.ssh/config
          # initialize kerberos
          export KRB5CCNAME=/var/tmp/crw-build_ccache
          kinit "crw-build/codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com@REDHAT.COM" -kt ''' + CRW_KEYTAB + '''
          klist # verify working

          function tag_and_push () {
              sudo podman tag $BREW_IMG quay.io/wto/$1
              sudo podman push quay.io/wto/$1
          }

          # $1 is the directory to cd into
          # $2 is the name of the image on quay
          function release () {
              echo "Starting $1 release process"

              # Store the log in the root
              rhpkg --verbose container-build --target=web-terminal-1.0-rhel-8-containers-candidate > ../tmp.log 2>&1
              NVR=$(cat ../tmp.log | grep -A1 "nvrs:" | tail -n 1) # e.g. web-terminal-operator-metadata-container-1.0.0-15
              TAG=$(echo $NVR | grep -o '[0-9]*.[0-9]*.[0-9]*-[0-9]*$') # e.g. 1.0.0-15

              cd .. # Go back out to the root where the projects are being cloned
              echo "Finished $1 release process. Brew NVR is: $NVR. Tag is $TAG"
              rm tmp.log

              BREW_IMG=registry-proxy.engineering.redhat.com/rh-osbs/$2:$TAG
              sudo podman pull $BREW_IMG

              # Push images to quay
              tag_and_push "$2:latest"
              tag_and_push "$2:$TAG"
              tag_and_push "$2:1.0.0"
          }
          '''

          if (params.BUILD_TOOLING) {
            sh BOOTSTRAP + '''
            git clone ssh://crw-build@pkgs.devel.redhat.com/''' + TOOLING_REPO  + '''.git web-terminal-tooling
            cd web-terminal-tooling
            git fetch -a
            git checkout ''' + SOURCE_BRANCH + '''
            release "web-terminal-tooling" "web-terminal-tooling"
            '''
          }
          
          if (params.BUILD_EXEC) {
            sh BOOTSTRAP + '''
            git clone ssh://crw-build@pkgs.devel.redhat.com/''' + EXEC_REPO  + '''.git web-terminal-exec
            cd web-terminal-exec
            git fetch -a
            git checkout ''' + SOURCE_BRANCH + '''
            release "web-terminal-exec" "web-terminal-exec"
            '''
          }

          if (params.BUILD_OPERATOR) {
            sh BOOTSTRAP + '''
            git clone ssh://crw-build@pkgs.devel.redhat.com/''' + OPERATOR_REPO  + '''.git web-terminal
            cd web-terminal
            git fetch -a
            git checkout ''' + SOURCE_BRANCH + '''
            release "web-terminal" "web-terminal-operator"
            '''
          }

          if (params.BUILD_OPERATOR_METADATA) {
            sh BOOTSTRAP + '''
            git clone ssh://crw-build@pkgs.devel.redhat.com/''' + OPERATOR_METADATA_REPO  + '''.git web-terminal-dev-operator-metadata
            cd web-terminal-dev-operator-metadata
            git fetch -a
            git checkout ''' + SOURCE_BRANCH + '''
            release "web-terminal-dev-operator-metadata" "web-terminal-operator-metadata"
            '''
          }
        }
      }
    }
  }
}
