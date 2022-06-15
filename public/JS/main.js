//Comment
function showComment() {
    var commentArea = document.getElementById("comment-area");
    commentArea.classList.remove("hide");
}


function upvote(Id) {
    fetch("http://127.0.0.1:8000/api/vote", {
        "headers": {
            "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
            "accept-language": "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7",
            "cache-control": "no-cache",
            "content-type": "application/x-www-form-urlencoded",
            "pragma": "no-cache",
            "sec-ch-ua": "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"102\", \"Google Chrome\";v=\"102\"",
            "sec-ch-ua-mobile": "?0",
            "sec-ch-ua-platform": "\"Windows\"",
            "sec-fetch-dest": "document",
            "sec-fetch-mode": "navigate",
            "sec-fetch-site": "same-origin",
            "sec-fetch-user": "?1",
            "upgrade-insecure-requests": "1"
        },
        "referrer": "http://127.0.0.1:8000/post?id="+Id,
        "referrerPolicy": "strict-origin-when-cross-origin",
        "body": "postId="+Id+"&vote=1",
        "method": "POST",
        "mode": "cors",
        "credentials": "include"
    }).then(() => {
        location.reload();
    });
}

function downvote(Id) {
    fetch("http://127.0.0.1:8000/api/vote", {
        "headers": {
            "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
            "accept-language": "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7",
            "cache-control": "no-cache",
            "content-type": "application/x-www-form-urlencoded",
            "pragma": "no-cache",
            "sec-ch-ua": "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"102\", \"Google Chrome\";v=\"102\"",
            "sec-ch-ua-mobile": "?0",
            "sec-ch-ua-platform": "\"Windows\"",
            "sec-fetch-dest": "document",
            "sec-fetch-mode": "navigate",
            "sec-fetch-site": "same-origin",
            "sec-fetch-user": "?1",
            "upgrade-insecure-requests": "1"
        },
        "referrer": "http://127.0.0.1:8000/post?id=" + Id,
        "referrerPolicy": "strict-origin-when-cross-origin",
        "body": "postId=" + Id + "&vote=-1",
        "method": "POST",
        "mode": "cors",
        "credentials": "include"
    }).then(() => {
        location.reload();
    });
}