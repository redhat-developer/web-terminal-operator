#!/bin/bash

# bootstrapping: if keytab is lost, upload to
# https://codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com/credentials/store/system/domain/_/
# then set Use secret text above and set Bindings > Variable (path to the file) as ''' + CRW_KEYTAB + '''
chmod 700 "${CRW_KEYTAB}" && chown "${USER}" "${CRW_KEYTAB}"
# create .k5login file
echo "crw-build/codeready-workspaces-jenkins.rhev-ci-vms.eng.rdu2.redhat.com@REDHAT.COM" > ~/.k5login
chmod 644 ~/.k5login && chown "${USER}" ~/.k5login
echo "pkgs.devel.redhat.com,10.19.208.80 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAplqWKs26qsoaTxvWn3DFcdbiBxqRLhFngGiMYhbudnAj4li9/VwAJqLm1M6YfjOoJrj9dlmuXhNzkSzvyoQODaRgsjCG5FaRjuN8CSM/y+glgCYsWX1HFZSnAasLDuW0ifNLPR2RBkmWx61QKq+TxFDjASBbBywtupJcCsA5ktkjLILS+1eWndPJeSUJiOtzhoN8KIigkYveHSetnxauxv1abqwQTk5PmxRgRt20kZEFSRqZOJUlcl85sZYzNC/G7mneptJtHlcNrPgImuOdus5CW+7W49Z/1xqqWI/iRjwipgEMGusPMlSzdxDX4JzIx6R53pDpAwSAQVGDz4F9eQ==
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

sudo yum install skopeo podman -y

#########################################################################
#
#                     Check out sources
#
#########################################################################
cd "${WORKSPACE}/sources" || exit
  git checkout --track origin/"${SOURCE_BRANCH}" || true
  git config user.email "jpinkney+web-terminal-release@gmail.com"
  git config user.name "Red Hat Web Terminal Release Bot"
  git config --global push.default matching
cd ..

# fetch sources to be updated
if [[ ! -d ${WORKSPACE}/targetdwn ]]; then git clone "ssh://crw-build@pkgs.devel.redhat.com/${DWNSTM_REPO}" targetdwn; fi
cd "${WORKSPACE}/targetdwn" || exit
  git checkout --track origin/"${DWNSTM_BRANCH}" || true
  git config user.email crw-build@REDHAT.COM
  git config user.name "CRW Build"
  git config --global push.default matching
cd ..

#########################################################################
#
#             Copy sources from upstream branch to dist-git
#
#########################################################################

# copy over the files
cp -r sources/* targetdwn/
