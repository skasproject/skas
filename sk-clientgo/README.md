
hash [password]
init <url>
login [user, [password]]   
logout
version
whoa-mi

// Admin func
user explain <user>
user create <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] [--password <password>] [--generatePassword] [--inputPassword]
user patch <user> [--email <email>] [--commonName <commonName>] [--uid <uid>] [--state <enabled|disabled>] [--comment <comment>] [--password <password>] [--generatePassword] [--inputPassword]
user bind <user> <group>
user unbind <user> <group>
user password <user>

