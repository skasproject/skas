
hash [password]
init <url>
login [user, [password]]   
logout
version
whoami

Admin func

    user describe <user> [--explain]

    user create <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] 
        [--password <password>] [--passwordHash <passwordHash>] [--generatePassword] [--inputPassword]

    user patch <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] 
        [--password <password>]  [--passwordHash <passwordHash>] [--generatePassword] [--inputPassword] [--create]

    user bind <user> <group> [--bindngName] [--strict] 
    
    user unbind <user> <group> [--bindngName] [--strict]


Use user patch to change a user password 

Following are redundant with kubectl. TODO ?
user list 
user get <user> [-o json|yaml]

token list
token delete <token>