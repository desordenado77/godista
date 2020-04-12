# godista
Dista in esperanto means remote

This is a tool I wrote to ease opening files hosted on my linux machine and shared with samba from my windows machine. It basically consists on a server running on the windows side, listening on a TCP/IP port that opens a file requested from the linux side. This allows me to open files for editing in Windows from the linux command line.

The same application works on both sides. I build it for windows doing 

GOOS=windows go build

and call it as a server with:

godista -s -f path_to_config_file

From the linux command line you can call godista the following way:

godista -c command_name -p command_parameters

If you install the application with:

godista -i

Aliases for all the defined commands will be created with the name dista<command_name>
