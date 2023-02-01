#!/bin/bash -p
FULLSCRIPT=$(readlink -f $0)
export ProgName=$(basename $FULLSCRIPT)

GHPCDIR=$HOME/dev/hpc-toolkit

sub_help () {
  cat <<EOF
Usage: $ProgName <deployment_name>.yaml <subcommand>

Deployment name:
  basename of local configuration yaml file

Subcommands:
  enable [-s]				Enable gcloud project, service-account, optional: add-iam-policy-binding

  create				Run ghcp create with --vars overriding blueprint yaml

  deploy				Terraform init, validate, apply commands

  destroy				Terraform destroy command

  redeploy				Terraform destroy command followed by deploy subcommand
EOF
}

function leading_tab_fmt {
  printf "___$1\n" | fmt -p"___" | sed -re 's/___/\t/'
}

# See
# https://stackoverflow.com/questions/5014632/how-can-i-parse-a-yaml-file-from-a-linux-shell-script
function parse_yaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
   }'
}

echoprint () {
  printf "\n------\n\t$1\n------\n\n"
}

terraform_cmd () {
  echoprint "Running terraform $1:"
  case $1 in
    apply|destroy)
      terraform -chdir=$DEPLOYMENT_NAME/primary $1 -auto-approve
      ;;
    *)
      terraform -chdir=$DEPLOYMENT_NAME/primary $1
      ;;
  esac
}

help_subcmd () {
  printf "$ProgName $1\n"
  leading_tab_fmt "$2"
}

sub_enable () {
  case "$1" in
    -h|--help|help)
      help_subcmd "enable [-s]" \
        "gcloud setup. If -s option present, add editor role for service account (not necessary for group accounts)"
      exit 1
      ;;
  esac

  while getopts "s" opt ; do
    case $opt in
      s)
        setup_service_account_editor_role=1
        ;;
      *)
        setup_service_account_editor_role=0
        ;;
    esac
  done

  set -x
  gcloud config set compute/zone $vars_zone || exit 1

  gcloud config set compute/region $vars_region || exit 1

  if [[ -v service_account_email ]] ; then
    gcloud iam service-accounts enable \
           --project=$project \
           $service_account_email || exit 1
  fi

  if ((setup_service_account_editor_role)) ;then
    gcloud projects add-iam-policy-binding $project \
      --member=serviceAccount:$service_account_email \
      --role=roles/editor || exit 1
  fi
  set +x
  }

sub_create () {
  case "$1" in
    -h|--help|help)
      help_subcmd create \
                  "Run ghpc create with extra vars setting"
      exit 0
      ;;
  esac

  $GHPCDIR/ghpc --version
  # Turn on command echo so we can see the commandline
  set -x
  eval $GHPCDIR/ghpc create -w $yamlfile $ghcp_vars
  # Then turn it off again
  set +x
  }

sub_destroy () {
  case "$1" in
    -h|--help|help)
      help_subcmd destroy \
                  "Terraform destroy command"
      exit 0
      ;;
  esac

  terraform_cmd destroy
}

sub_setup_ssh () {
  case "$1" in
    -h|--help|help)
      help_subcmd "setup_ssh [logfile]" \
                  "Setup ssh keys and config stanzas using output of deploy \
subcommand. First argument is the deploy logfile. \
If none is provided, the latest ${DEPLOYMENT_NAME}_*deploy_*.log \
file will be used."
      exit 0
      ;;
    *.log)
      export sshlogfile=$1
      ;;
    *)
      export sshlogfile=$(ls -1 -rt ${DEPLOYMENT_NAME}_*deploy_*.log | tail -1)
      ;;
  esac
  printf "\nsetup_ssh subcommand: sshlogfile set to $sshlogfile\n\n"

  : > setup_ssh_filelist

  grep name_prefix $sshlogfile | grep -v known | egrep -o '[^"]+-(login|controller)' | \
    while read name_prefix ; do
      gcloud compute instances list --project=$project | egrep "$name_prefix" | awk '{print $1, $2}' | \
        while read hka zone ; do
          export gcloud_ssh="gcloud compute ssh $hka --project=$project --zone=$zone --tunnel-through-iap"

          ssh_filename="ssh_${hka}.sh"
          printf "#!/bin/sh\n$gcloud_ssh \$*\n" > $ssh_filename
          chmod +x $ssh_filename
          echo $ssh_filename >> setup_ssh_filelist

          echo "Running './$ssh_filename --command=whoami' to set up keys for $hka on this host"
          echo "(that script contains the command '${gcloud_ssh} \$*')"
          remote_user="$(./$ssh_filename --command=whoami 2>/dev/null)"
          echo "remote_user set to $remote_user"

          hostalias="$name_prefix"
          if [ "$hostalias" != "$hka" ] ; then
            hostalias="$hostalias $hka"
          fi


          mkdir -p config.d
          ssh_config_filename="config.d/${hka}"
          echo "$ssh_config_filename" >> setup_ssh_filelist
          cat > $ssh_config_filename <<-EOF
	Host $hostalias
	   User				${remote_user}
	   HostKeyAlias			$hka
	   Hostname			$hka.$project.$zone.compute.gcp
	
	EOF

          cat <<-EOF
	-----------------------------------
	This host has ssh keys set up to use ssh from the commandline.
	
	Before using commandline ssh from other hosts, connect via the gcloud connect ssh
	command saved into the ./$ssh_filename shell script, e.g.:
	
		./$ssh_filename --command=whomami
	
	An ssh configuration stanza has been saved into ./$ssh_config_filename
	
	Insert ./$ssh_config_filename into your ~/.ssh/config.d directory on this and other hosts.
	Then add the line 'Include "~/.ssh/config.d/*"' at the top of ~/.ssh/config .
	Its contents are
	
	EOF
          sed -re 's/^/\t/' $ssh_config_filename
      done
    done

  gcp_compute_filename="config.d/zzz_match_host_compute_gcp"
  echo "$gcp_compute_filename" >> setup_ssh_filelist
  echo "------"
  cat <<-EOF
	
	The ssh config stanzas in config.d rely on the stanza saved to $gcp_compute_filename being found
	later in ~/.ssh/config.d/*. Its contents look like this:
	
	EOF

  cat > $gcp_compute_filename <<-EOF
	# Pseudo-host format for GCP compute VM instances:
	#    instance.project.zone.compute.gcp
	#
	# Host short-name instance-name
	#     User           ldap_google_com
	#     HostKeyAlias   instance-name
	#     HostName       instance-name.project.zone.compute.gcp
	#
	# Add additional stanzas for different gcloud groups
	#
	# Be sure you set the HostKeyAlias to the instance name.
	Match Host *.*.*.compute.gcp
	    IdentityFile                ~/.ssh/google_compute_engine
	    UserKnownHostsFile          ~/.ssh/google_compute_known_hosts
	    IdentitiesOnly              yes
	    CheckHostIP                 no
	    ProxyUseFdpass              no
	    ProxyCommand                gcloud compute start-iap-tunnel %k %p --listen-on-stdin --project=$(echo %h| cut -d. -f2) --zone=$(echo %h| cut -d. -f3)
	
	EOF

  sed -re 's/^/\t/' $gcp_compute_filename

  printf "\nNew files created by setup_ssh subcommand:\n\n"
  ls -lrt $(cat setup_ssh_filelist) | sed -re 's/^/\t/'
  /bin/rm -f setup_ssh_filelist
  echo "------"
}

sub_deploy () {
  case "$1" in
    -h|--help|help)
      help_subcmd destroy \
                  "Terraform init, validate, apply commands"
      exit 0
      ;;
  esac

  terraform_cmd init && \
  terraform_cmd validate && \
  terraform_cmd apply && \
  echoprint "Successful completion" && \
  sub_setup_ssh $logfile
}

sub_redeploy () {
  case "$1" in
    -h|--help|help)
      help_subcmd destroy \
                  "Run destroy subcmd, then deploy subcmd"
      exit 0
      ;;
  esac

  sub_destroy && sub_deploy
}

if (($# < 2)) ; then
  sub_help
  exit 1
fi

ProgName=$(basename $0)
export DEPLOYMENT_NAME=$(basename $1 .yaml)
input_yaml=$(dirname $1)/${DEPLOYMENT_NAME}.yaml

echo DEPLOYMENT_NAME = $DEPLOYMENT_NAME

if [[ -r $input_yaml ]]  ; then
  eval $(parse_yaml $input_yaml)
  if [[ -z "$project" ]] ; then
    echo "$input_yaml does not contain a 'project:' specification"
    sub_help
    exit 1
  fi
  ghcp_vars="$(parse_yaml $input_yaml | grep '^vars_'| sed -e 's/vars_/--vars /g' -e 's/\"//g')"
  shift
else
  echo "$input_yaml is not a yaml file"
  sub_help
  exit 2
fi

subcommand="$1"
case $subcommand in
  ""|"-h"|"--help"|"help")
    sub_help
    exit 0
    ;;
  *)
    shift
    export logfile=${DEPLOYMENT_NAME}_${subcommand}_$(date -Iminutes).log
    set -o pipefail
    sub_${subcommand} ${1+"$@"} 2>&1 | tee $logfile
    return_code=$?
    set +o pipefail
    if (( $return_code == 127 )) ; then
      echo "Error: '$subcommand' is not a known subcommand." >&2
      echo "       Run '$ProgName --help' for a list of known subcommands." >&2
    else
      echo "$subcommand stdout and stderr saved to $logfile"
    fi
    exit $return_code
    ;;
esac
