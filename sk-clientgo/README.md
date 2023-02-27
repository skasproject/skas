
hash [password]
init <url>
login [user, [password]]   
logout
version
whoa-mi

Admin func

    user explain <user>

    user create <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] 
        [--password <password>] [--passwordHash <passwordHash>] [--generatePassword] [--inputPassword]

    user patch <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] 
        [--password <password>]  [--passwordHash <passwordHash>] [--generatePassword] [--inputPassword] [--create]

    user bind <user> <group> [--noError]
    
    user unbind <user> <group> [--noError]
    
    user password <user>

Following are redundant with kubectl. TODO ?
user list 
user get <user> [-o json|yaml]

token list
token delete <token>