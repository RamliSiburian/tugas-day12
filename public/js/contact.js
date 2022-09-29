function submitData() {
    let name = document.getElementById("name").value
    let email = document.getElementById("email").value
    let phoneNumber = document.getElementById("phone-number").value
    let subject = document.getElementById("subject").value
    let message = document.getElementById("message").value
    let emailReceiver = "ramlisiburian1111@gmail.com";

    if (name == "" || email == "" || phoneNumber == "" || subject == "" || message == "") {
        return alert("field tidak boleh kosong")
    }


    let a = document.createElement('a')
    // a.href = `mailto:${emailReceiver}?subject=${subject}&body=${message}`  

    a.href = `https://mail.google.com/mail/?view=cm&fs=1&to=${emailReceiver}&su=${subject}&body=Hello my name is ${name}, ${message} , ${phoneNumber}`
    a.click();
}