
# Getting started

## What is hero ?

From the definition of oauth 2 above, there is example of providers that is Facebook,GitHub and DigitalOcean. `hero` enables you to become an oauth 2 provider just like Facebook ,GitHub and DigitalOcean.

This means, hero provides user accounts and also allows  limited access to the accounts.Unlike the other options, with hero you have full control of everything.

Hero is a commadline application.  It also offers a library that you can use to compose your own version of oauth 2 provider.


# Installation

## From source

Prerequisite
* go(Golang) 1.5+

Go get the project to install it

	go get github.com/gernest/hero/cmd/hero

This will create the hero binary for you.

## Precompiled
Alternatively you can download precompiled binaries for your favorite operating system.

[COMING SOON]

## Usage
You need a configuration file  in order to run `hero` server. The format supported is json. Hero comes with a command to bootstrap a configuration file with default settings, you can use it to customize the values as you fancy.I will explain the configurable details in a moment.


### Step 1 Configure

To generate default configuration file, run the following command,

	hero genconf [Path to configuration file goes here e.g config.json]

### Step 2: Run

Run the service

	hero --mingrate server [config_file_path]

where `config_file_path` is the path to the configuration file, it doisnt matter if the path is relative or absolute.

It is wise to add the `--migrate` flag if you are running for the first time so as to prepare the database.


You should see this on your `stdout`

	running migrations...done 
	starting hero service at  http://localhost:8090 
