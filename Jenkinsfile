pipeline {
agent {
   // dockerfile true
   docker {
    image 'golang:alpine'
    args '-v /tmp/nodetest:/root/nodetest'
   }
}

 environment {
        CC = 'clang'
        REGISTRY_KEY=credentials('16e666af-c002-483d-88ce-7a9c90b06ddd')
    }

 options {
        timeout(time: 1, unit: 'HOURS')
    }

  parameters {
            string(name: 'PERSON', defaultValue: 'Mr Jenkins', description: 'Who should I say hello to?')
        }


 stages {

      // Check out the code from Gerrit
         stage('Check out the code') {

             environment {
                 AN_ACCESS_KEY = credentials('16e666af-c002-483d-88ce-7a9c90b06ddd')
             }

                agent {
                        docker {
                            image 'jermine/git'
                              }
                    }
                    steps {
                        echo '$AN_ACCESS_KEY'
                        echo 'Check out the code'
                    }
                }

     // Used SonarQube to scan codes
     stage('SonarQube to Analysis code') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'SonarQube to Analysis code'
         }
     }

     // Run unit test
     stage('Run unit test') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'Run unit test in current branch code'
         }
     }


     // Build docker images
     stage(' Build docker images for this branch') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo ' Build docker images for this branch'
         }
     }


     // Push docker images to registry
     stage(' Push docker images to registry') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'Push docker images to registry'
         }
     }

    // Deploy from registry
     stage('Deploy from registry') {
                when {
                      allOf {
                            branch 'master'
                            environment name: 'DEPLOY_TO', value: 'production'
                        }
                    }
                    steps {
                        echo 'Deploy from registry'
                    }
         }

     // Run E2E test , maybe API or Web UI or Monkey tools .This stage should be run on windows node will better.(If app has been UI)
     stage('Run E2E test') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'Run E2E test'
         }
     }


     // Update the label with Verified.
     stage('Update the label with Verified') {
         when {
             allOf {
                 branch 'master'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'Update the label with Verified'
         }

         // To notified repositroy keeper by DingDing IM
         post {
             always {
                 echo 'I will always say Hello again!'
             }
             success{
                 echo 'The Gerrit already update the label, You can merge the code to repo.'
             }

             failure {
                 mail to: "Jermine.hu@qq.com", subject: 'The Pipeline failed :('
             }
         }

     }

     // Deploy from registry to production by k8s api in helm
     stage('Deploy from registry to production') {
         when {
             allOf {
                 branch 'release'
                 environment name: 'DEPLOY_TO', value: 'production'
             }
         }
         steps {
             echo 'Deploy from registry to production'
         }

         // To notified repositroy keeper by DingDing IM
         post {
             always {
                 echo 'I will always say Hello again!'
             }
             success{
                 echo 'Deploy from registry to production successfully.'
             }

             failure {
                 mail to: "Jermine.hu@qq.com", subject: 'Deploy from registry to production was failure.'
             }
         }

     }

         script {
                    def browsers = ['chrome', 'firefox']
                    for (int i = 0; i < browsers.size(); ++i) {
                        echo "Testing the ${browsers[i]} browser"
                    }
                }

    }

     post {
            always {
                echo 'Pipeline in the end'
            }
        }


}