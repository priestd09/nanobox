// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

//
package vagrant

import (
	"fmt"
	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox/config"
	"github.com/nanobox-io/nanobox/util"
	engineutil "github.com/nanobox-io/nanobox/util/engine"
	fileutil "github.com/nanobox-io/nanobox/util/file"
	"os"
	"path/filepath"
)

// Init
func Init() {

	// create Vagrantfile
	vagrantfile, err := os.Create(config.AppDir + "/Vagrantfile")
	if err != nil {
		config.Fatal("[commands/init] ioutil.WriteFile() failed", err.Error())
	}
	defer vagrantfile.Close()

	// create synced folders
	synced_folders := fmt.Sprintf("nanobox.vm.synced_folder \"%s\", \"/vagrant/code/%s\"", config.CWDir, config.Nanofile.Name)

	// attempt to parse the boxfile first; we don't want to create an app folder
	// if the app isn't able to be created
	boxfile := config.ParseBoxfile()

	// if an custom engine path is provided, add it to the synced_folders
	if enginePath := boxfile.Build.Engine; enginePath != "" {
		if _, err := os.Stat(enginePath); err == nil {

			//
			base := filepath.Base(enginePath)

			//
			appEngineDir := filepath.Join(config.AppDir, base)
			if _, err := os.Stat(appEngineDir); err != nil {
				if err := os.Mkdir(appEngineDir, 0755); err != nil {
					config.Fatal("[commands/init] os.Mkdir() failed", err.Error())
				}
			}

			synced_folders += fmt.Sprintf("\n    nanobox.vm.synced_folder \"%s\", \"/vagrant/engines/%s\"", appEngineDir, base)

			//
			whatever := &struct {
				Overlays []string `json:"overlays"`
			}{}

			//
			enginefile := filepath.Join(enginePath, "./Enginefile")

			// if no engine file is found just return
			if _, err := os.Stat(enginefile); err != nil {
				fmt.Println("Enginefile not found, Exiting... ")
				os.Exit(1)
			}

			// parse the ./Enginefile into the new release
			if err := config.ParseConfig(enginefile, whatever); err != nil {
				fmt.Printf("Nanobox failed to parse your Enginefile. Please ensure it is valid YAML and try again.\n", err.Error())
				config.Log.Error("[commands/engine/publish] http.Get() failed", err.Error())
				os.Exit(1)
			}

			fmt.Printf("WHATEVER! %#v\n", whatever)

			// iterate through each overlay fetching it and adding it to the list of 'files'
			// to be tarballed
			for _, overlay := range whatever.Overlays {

				fmt.Println("OVERLAY!", overlay)

				// extract a user and archive (desired engine) from args[0]
				user, archive := engineutil.ExtractArchive(overlay)

				// extract an engine and version from the archive
				e, version := engineutil.ExtractEngine(archive)

				//
				res, err := engineutil.GetEngine(user, e, version)
				if err != nil {
					config.Fatal("[commands/engine/publish] http.Get() failed", err.Error())
				}
				defer res.Body.Close()

				//
				switch res.StatusCode / 100 {
				case 2, 3:
					break
				case 4, 5:
					os.Stderr.WriteString(stylish.ErrBullet("Unable to fetch '%v' overlay, exiting...", e))
					os.Exit(1)
				}

				//
				if err := fileutil.Untar(appEngineDir, res.Body); err != nil {
					config.Fatal("[commands/engine/publish] file.Untar() failed", err.Error())
				}
			}

			abs, err := filepath.Abs(enginePath)
			if err != nil {
				fmt.Println("BONK!", err)
			}

			fmt.Println("ABS!", abs)

			// pull the remainin engine files over
			for _, f := range []string{"bin", "Enginefile", "lib", "templates", "files"} {

				path := filepath.Join(abs, f)

				//
				if _, err := os.Stat(path); err != nil {
					continue
				}

				// not handling error here because an error simply means the file doesn't
				// exist and therefor wont be copied
				if err := fileutil.Copy(path, appEngineDir); err != nil {
					config.Fatal("[commands/engine/publish] file.Copy() failed", err.Error())
				}
			}
		}
	}

	//
	// nanofile config
	//
	// create nanobox private network and unique forward port
	network := fmt.Sprintf("nanobox.vm.network \"private_network\", ip: \"%s\"", config.Nanofile.IP)
	sshport := fmt.Sprintf("nanobox.vm.network :forwarded_port, guest: 22, host: %v", util.StringToPort(config.Nanofile.Name))

	//
	provider := fmt.Sprintf(`# VirtualBox
    nanobox.vm.provider "virtualbox" do |p|
      p.name = "%v"

      p.customize ["modifyvm", :id, "--cpuexecutioncap", "%v"]
      p.cpus = %v
      p.memory = %v
    end`, config.Nanofile.Name, config.Nanofile.CPUCap, config.Nanofile.CPUs, config.Nanofile.RAM)

	//
	// insert a provision script that will indicate to nanobox-server to boot into
	// 'devmode'
	var devmode string
	if config.Devmode {
		fmt.Printf(stylish.Bullet("Configuring vm to run in 'devmode'"))

		devmode = `# added because --dev was detected; boots the server into 'devmode'
    config.vm.provision "shell", inline: <<-DEVMODE
      echo "Starting VM in dev mode..."
      mkdir -p /mnt/sda/var/nanobox
      touch /mnt/sda/var/nanobox/DEV
    DEVMODE`
	}

	//
	// write to Vagrantfile
	vagrantfile.Write([]byte(fmt.Sprintf(`
################################################################################
##                                                                            ##
##                                   ***                                      ##
##                                *********                                   ##
##                           *******************                              ##
##                       ***************************                          ##
##                           *******************                              ##
##                       ...      *********      ...                          ##
##                           ...     ***     ...                              ##
##                       +++      ...   ...      +++                          ##
##                           +++     ...     +++                              ##
##                       \\\      +++   +++      ///                          ##
##                           \\\     +++     ///                              ##
##                                \\     //                                   ##
##                                   \//                                      ##
##                                                                            ##
##                    _  _ ____ _  _ ____ ___  ____ _  _                      ##
##                    |\ | |__| |\ | |  | |__) |  |  \/                       ##
##                    | \| |  | | \| |__| |__) |__| _/\_                      ##
##                                                                            ##
## This file was generated by nanobox. Any modifications to it may cause your ##
## nanobox VM to fail! To regenerate this file, delete it and run             ##
## 'nanobox init'                                                             ##
##                                                                            ##
################################################################################

# -*- mode: ruby -*-
# vi: set ft=ruby :

#
Vagrant.configure(2) do |config|

  # add the boot2docker user credentials to allow nanobox to freely ssh into the vm
  # w/o requiring a password
  config.ssh.shell = "bash"
  config.ssh.username = "docker"
  config.ssh.password = "tcuser"

  config.vm.define :'%s' do |nanobox|

    ## Set the hostname of the vm to the app domain
    nanobox.vm.provision "shell", inline: <<-SCRIPT
      sudo hostname %s
    SCRIPT

    ## Wait for nanobox-server to be ready before vagrant exits
    nanobox.vm.provision "shell", inline: <<-WAIT
      echo "Waiting for nanobox server..."
      while ! nc -z 127.0.0.1 1757; do sleep 1; done;
    WAIT

    ## box
    nanobox.vm.box     = "nanobox/boot2docker"


    ## network
    %s
    %s


    ## shared folders

    # disable default /vagrant share (overridden below)
    nanobox.vm.synced_folder ".", "/vagrant", disabled: true

    # add nanobox shared folders
    nanobox.vm.synced_folder "~/.ssh", "/mnt/ssh"
    %s


    ## provider configs
    %s

    ## wait for the dhcp service to come online
    nanobox.vm.provision "shell", inline: <<-WAIT
      attempts=0
      while [ ! -f /var/run/udhcpc.eth1.pid && $attempts -lt 30 ]; do
        let attempts++
        sleep 1
      done
    WAIT

    # kill the eth1 dhcp server so that it doesn't override the assigned ip when
    # the lease is up
    nanobox.vm.provision "shell", inline: <<-KILL
      if [ -f /var/run/udhcpc.eth1.pid ]; then
        echo "Killing eth1 dhcp..."
        kill -9 $(cat /var/run/udhcpc.eth1.pid)
      fi
    KILL

    %s

  end
end`, config.Nanofile.Name, config.Nanofile.Domain, network, sshport, synced_folders, provider, devmode)))
}
