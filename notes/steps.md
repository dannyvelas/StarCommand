1. create ansible inventory file with hosts and its VMs. VMs have proxyjump set to true so that later on, ansible playbooks can be run on them
2. ansible-playbook bootstrap-server.yml -u root # runs for all hosts added in inventory in step#1
3. ansible-playbook setup-host.yml # runs for all hosts added in inventory in step#1
   for each host {
4.   add {{.HOST}} to ~/.ssh/config (if not already there)

5.   # OVN: first host creates central DB, subsequent hosts join as chassis
     if existing ovn cluster {
       ovn chassis add {{.HOST}}
     } else {
       ovn init {{.HOST}}
     }

6.   # k3s server: first host initializes cluster, subsequent hosts join
     if existing k3s cluster {
       k3s join {{.HOST}}
     } else {
       k3s init {{.HOST}}
     }

7.   cd terraform && terraform init && terraform apply
8.   ansible-playbook bootstrap-server.yml --limit vms -u root
9.   ansible-playbook setup-vm.yml
     for vm in host.vms {
10.     add vm to ~/.ssh/config (if not already there)
       # k3s agent: VM joins the cluster as a worker
11.    k3s agent join vm
     }

12.  # incus: first host creates cluster, subsequent hosts join
     if existing incus cluster {
       incus cluster join {{.HOST}}
     } else {
       incus remote add my-homelab
       incus remote switch my-homelab
     }
   }
14. incus list
