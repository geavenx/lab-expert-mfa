# Instructions
- Change the host to the URL of your lab
- Run the program with '''go run main.go'''
- When session founded, copy them
- Open burpsuite and open any http request to the lab, send it to repeater
- Modify the request, to be a GET to the /my-account endpoint
- Paste the session that you copied into the cookie, like this: 'Cookie: Session=your session here!'
- Now see the response in the browser and you're done.