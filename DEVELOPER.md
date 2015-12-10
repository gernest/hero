# Why you should contribute to hero?

Be a hero, fight against poverty in style. I have a vision, a way in which I can somehow make a difference in the fight against povert( Trust me, I'm in deep shit of poverty) in my country(Tanzania) with the power of open source and creative thinking.

The people who are entitled to make decisions( at the lowest level) ni my country , they have no idea what they are doing. So with the assist of data science, open source, science fiction and collaboration I hope one day a clone of me(  yes the other version of me somewhere in the distant future) will have the chance to do what he wants to do with his full capacity.



# Developing
Okay, time to be part of `heroes` squad. Before you leave this section please don't forget to star the project.

__NOTE__ : For non trivial stuffs like typo fixes just use the [GitHub interface](https://help.github.com/articles/github-flow-in-the-browser/) and don't bother with the boring recomendation below. If you are not so sure, better open an issue.

## Boring workflow( for the newbies who haven't contributed to complex golang project yet)
Prerequisite

* A working go(Golang) environment
* go 1.5+
* database (mysql,postgres or foundation)
* make
* A brave heart

I am using a linux machine, so this is a linux based setup.

Fork the project on github. Assuming your github username is `foo`. After forking you should have the `hero` fork on your github repositories.

get the project

	go get -v -u github.com/gernest/hero/cmd/hero


cd to the installed package

	cd $GOPATH/src/github.com/gernest/hero

add your fork as remote call it  heroes

	git remote add heroes URL_TO_YOUR_FORK

where URL_TO_YOUR_FORK is the github url to your fork of hero e.g https://github.com/foo/hero


clean the development environment

	make clean
	

Install dependencies

	make deps

create a branch that you will be hacking on, let us call it fix

	git checkout -b fix
	

Export the database connection setup. We need dialect and the connection url. Default values are
`DB_CONN=postgres://postgres:postgres@localhost/hero_test?sslmode=disable` and `DB_DIALECT=postgres`

You can create your database anywhere, just make sure it is available and supported. For now mysql, postgresql and foundation are supported.

	export DB_CONN=connection_url_to_my database
	export DB_DIALECT=my_database_dialect
	
fireup the goconvey test runner. This will rebuild and run the tests whenever any file changes, and you will see the convrage and the tests run on your default browser.

	make convey
	
Then hack on your fix branch.

If you are done hacking and the tests passed. Do some vetting and linting by running the following command.

	make check

Well if you are done and the check passes.

commit your changes, and push to your fork( our case heroes)

 	git commit -m -a
	
	git push heroes fix
	

Go to your github account and on your hero fork, open a pull request for your fix branch. I will review the pull request and merge it if the fix does what it says It can do.

## Running development server
If you are doing the front end or whatever and want to run the development server. First you should have somehow completed the boring recomendation above.

Then run the following command

	make server

The above command will start the development server. Database migrations will be done Before the server is started.

The configuration file used by this server is at the root of this respository [config_dev.json](config_dev.json)


Then you can view the home page by visiting [http://localhost:8090](http://localhost:8090)


# Running the hero demo.
There is a simple demo bundled with this repository at [hero/demo](hero/demo). 

To run the demo just do this command

	make demo

And visit <http://localhost:8001/login> to do oauth login.