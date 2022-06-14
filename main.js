//NavBar
function hideIconBar(){
    var iconBar = document.getElementById("iconBar");
    var navigation = document.getElementById("navigation");
    iconBar.setAttribute("style", "display:none;");
    navigation.classList.remove("hide");
}

function showIconBar(){
    var iconBar = document.getElementById("iconBar");
    var navigation = document.getElementById("navigation");
    iconBar.setAttribute("style", "display:block;");
    navigation.classList.add("hide");
}

//Comment
function showComment(){
    var commentArea = document.getElementById("comment-area");
    commentArea.classList.remove("hide");
}

//Reply
function showReply(){
    var replyArea = document.getElementById("reply-area");
    replyArea.classList.remove("hide");
}



function hideContainer(){
    var container = document.getElementById("container");
    container.classList.add("hide");
}
function showContainer(){
    var container = document.getElementById("container");
    container.classList.remove("hide");
}

function fasterPreview( uploader ) {
    if ( uploader.files && uploader.files[0] ){
        $('#profileImage').attr('src',
            window.URL.createObjectURL(uploader.files[0]) );
    }
}


