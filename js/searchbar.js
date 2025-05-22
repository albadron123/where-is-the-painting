const ADD_NEW_AUTHOR_STR = '<li><a href="/register_author">Добавить автора</a></li>'
//when we are adding new users we can't create new
const ADD_NEW_USER = ``

function search() {
    console.log("hi!")
}

function search_unfocus() {
    var node = document.getElementById('search-results');
    setTimeout(() => {node.innerHTML=''}, 100);
}


function search_focus(default_option) {
    var node = document.getElementById('search-results');
    node.innerHTML = default_option
}
