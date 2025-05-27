//=====================SUPPORT-FUNCTIONS-AKA-C-STYLE-INCLUDE-OF-UTILS======================================
function formHighlightedSubstring(initial, toFind)
{
    console.log(initial)
    console.log(toFind)
    index = initial.toLowerCase().indexOf(toFind.toLowerCase())
    return `${initial.substring(0,index)}<b style="background-color:yellow">${initial.substring(index, index+toFind.length)}</b>${initial.substring(index+toFind.length, initial.length)}`
}
//===============================END-OF-INCLUDE-SECTION====================================================


var GLOBAL_museumId = 0
window.onload=init;
async function init()
{    
    var node = document.getElementById('get-museum-id');
    GLOBAL_museumId = node.innerHTML;

    form = document.getElementById('form');
    form.addEventListener('submit', registerPainting);
}

async function registerPainting(event) 
{
    event.preventDefault();

    updateInfo = document.getElementById(`update-info`);

    //sanity check
    paintingName = document.getElementById(`name`);
    creationYear = document.getElementById(`creation-year`);
    whereToFind = document.getElementById(`where-to-find`);
    authorId = document.getElementById(`chosen-author-id`);
    fileInput = document.getElementById("fileinput");
    if(paintingName.value == '')
    {
        updateInfo.innerHTML = "Название не указано"
        return
    }
    if(authorId == null)
    {
        updateInfo.innerHTML = "Автор не выбран"
        return
    }
    if(fileInput.files.length != 1)
    {
        updateInfo.innerHTML = "Файл не выбран"
        return
    }

    yearValid = (parseInt(creationYear.value)!=NaN)
    data = {
        title: paintingName.value,
        creation_year: {
            Int32: parseInt(creationYear.value),
            Valid: yearValid,
        }, 
        where_to_find: whereToFind.value,
        author_id: parseInt(authorId.innerHTML),
    };
    formData = new FormData()
    jsonPart = JSON.stringify(data)
    filePart = fileInput.files[0]
    formData.append('json', jsonPart);
    formData.append('file', filePart);
    const response = await fetch(window.location.href, {
        method: 'POST',
        credentials: 'include',
        body: formData
    })
    if (!response.ok)
    {
        updateInfo.innerHTML = `что-то пошло не так: ${response.statusText}`;
    } 
    else
    {
        window.location.href = "/success";
    }
}

async function searchAuthors() {
    var node = document.getElementById('search-results');
    const query = document.getElementById("searchbar").value
    if(query == "")
    {
        node.innerHTML = 
        `
        <li><a href="/register_author">Добавить автора</a></li>
        `
        return
    }
    const responce = await fetch(`/authors_${query}`)
    if (!responce.ok)
    {
        //check for errors later
        node.innerHTML = `<li><a href="/register_author">Добавить автора</a></li>`
        return;
    }
    const data = await responce.json()
    innerHypertext = ''
    for (let i = 0; i < data.length; i++)
    {
        innerHypertext += 
        `
        <li class="search-res">
            <b>${formHighlightedSubstring(data[i].name, query)}</b><button onclick="chooseAuthor(${data[i].id},'${data[i].name}')">выбрать автора</button><hr>
        </li>
        `
    }
    innerHypertext += `<li><a href="/register_author">Добавить автора</a></li>`
    console.log(innerHypertext)
    console.log(data.length)
    node.innerHTML = innerHypertext;
}


function chooseAuthor(authorId, authorName)
{
    node = document.getElementById('search-view');
    node.innerHTML = `Автор: <div id="chosen-author-id" style="display:none;">${authorId}</div>${authorName}<button onclick="unchooseAuthor()">Сменить автора</button>`
}

function unchooseAuthor() 
{
    node = document.getElementById('search-view');
    node.innerHTML = `<input id="searchbar" oninput="searchAuthors()" onfocusout="search_unfocus()" onfocus="search_focus(ADD_NEW_AUTHOR_STR)">
                      <ul id="search-results"></ul>`
}