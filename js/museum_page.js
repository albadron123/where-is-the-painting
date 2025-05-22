var GLOBAL_museumId = 0
var GLOBAL_search_id = 0
var GLOBAL_query = ''

window.onload=init;
function init()
{
    var node = document.getElementById('get-museum-id');
    GLOBAL_museumId = node.innerHTML
}

async function search_users() {
    var node = document.getElementById('search-results');
    const query = document.getElementById("searchbar").value
    if(query == "")
    {
        node.innerHTML = 
        `
        <li>Введите что-нибудь...</li>
        `
        return
    }
    const responce = await fetch(`/museum${GLOBAL_museumId}/search_users_${query}`)
    if (!responce.ok)
    {
        //check for errors later
        node.innerHTML = `Ничего не найдено.`
        return;
    }
    const data = await responce.json()
    innerHypertext = ''
    for (let i = 0; i < data.length; i++)
    {
        innerHypertext += 
        `
        <li class="search-res">
            <b>${data[i].login}</b><button onclick="addUserToMuseum(${data[i].id},'${data[i].login}')">добавить</button><hr>
        </li>
        `
    };
    if(data.length == 0) 
    {
        innerHypertext = `Ничего не найдено.`
    }
    console.log(innerHypertext)
    console.log(data.length)
    node.innerHTML = innerHypertext;
}

async function addUserToMuseum(userId, userName)
{
    usersList = document.getElementById(`users-list`);
    updateInfo = document.getElementById(`update-info`);
    //isAdmin = checkbox.checked;
    data = {
        user_id : userId,
        is_admin : false,
    };
    const responce = await fetch(`http://localhost:8080/museum${GLOBAL_museumId}/rights`,
    {
        method: "POST",
        credentials: 'include',
        body: JSON.stringify(data),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });
    if (!responce.ok)
    {
        rnd = Math.floor(Math.random() * 3)+5;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">Добавить пользователя не удалось.</i>`
        checkbox.checked = !isAdmin;
        return;
    }
    else
    {
        usersList.innerHTML += `<li id="user${userId}">
                                    admin rights:<input type="checkbox" id="cb${userId}" onchange="checkRight(${userId})">
                                    ${userName}
                                    <button onclick="deleteUser(${userId})">удалить</button>
                                </li>`;
        rnd = Math.floor(Math.random() * 5)+3;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">Пользователь добавлен.</i>`
    }
}



function pressSearchButton()
{
    GLOBAL_search_id = 0
    GLOBAL_query = document.getElementById("big-searchbar").value
    searchPaintingsInMuseum(GLOBAL_search_id, GLOBAL_query)
}

function pressMore()
{
    GLOBAL_search_id += 1
    searchPaintingsInMuseum(GLOBAL_search_id, GLOBAL_query)
}

async function searchPaintingsInMuseum(pageId, query) {
    var node = document.getElementById('paintings-register');
    var morePlaceholder = document.getElementById('morePlaceholder');
    if(query == "")
    {
        node.innerHTML = 
        `
        Введите что-нибудь, чтобы посмотреть регистр картин музея
        `
        return
    }

    const responce = await fetch(`http://localhost:8080/museum${GLOBAL_museumId}/paintings_${query}/page${pageId}`,
    {
        method: "GET",
        credentials: 'include',
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });

    if (!responce.ok)
    {
        //may be fix the problem if it exists?
        return;
    }
    const data = await responce.json()
    console.log(data.length)
    if (pageId == 0) 
    {
        innerHypertext = ''
    }
    for (let i = 0; i < data.length; i++)
    {
        index = data[i].title.toLowerCase().indexOf(query.toLowerCase())
        innerHypertext += 
        `
        <li class="search-res" id="p${data[i].id}">
            "${data[i].title.substring(0,index)}<b style="background-color:yellow">${data[i].title.substring(index, index+query.length)}</b>${data[i].title.substring(index+query.length, data[i].title.length)}(${data[i].creation_year})"</b>
            автор: ${data[i].author_name}
            <button onclick="deletePainting(${data[i].id})">удалить</button>
            <hr>
        </li>
        `;
        console.log(data[i])
        /*
        ссылка на редактирование: <a href="#">редактировать</a> 

        АВТОР: <b>${data[i].author_name}</b><br>
                МУЗЕЙ: <b><a>${data[i].museum_name}</b></a><br>
                ГДЕ НАЙТИ: <b>${where_to_find}</b>
        */
    }
    if(data.length == 10)
    {
        morePlaceholder.innerHTML = `<button onclick="pressMore()">Загрузить ещё</button>`;
    }
    else 
    {
        morePlaceholder.innerHTML = ``;
        if(data.length == 0 && pageId == 0)
        {
            innerHypertext = `Картин с таким названием нет.`;
        }
    }
    console.log(innerHypertext);
    console.log(data.length);
    node.innerHTML = innerHypertext;
    //console.error(museum_id);
}


async function deletePainting(painting_id)
{
    var node = document.getElementById(`p${painting_id}`);
    const responce = await fetch(`http://localhost:8080/painting${painting_id}/delete_painting`,
    {
        method: "DELETE",
        credentials: 'include',
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });
    if (!responce.ok)
    {
        console.error("Not ok!")
        return;
    }
    else
    {
        console.error("Ok!")
        node.innerHTML = `<b style="color:grey">Картина удалена<b><hr>`
    }
}


async function checkRight(userId) 
{
    checkbox = document.getElementById(`cb${userId}`);
    updateInfo = document.getElementById(`update-info`);
    isAdmin = checkbox.checked;
    data = {
        user_id : userId,
        is_admin : isAdmin,
    };
    const responce = await fetch(`http://localhost:8080/museum${GLOBAL_museumId}/rights`,
    {
        method: "PUT",
        credentials: 'include',
        body: JSON.stringify(data),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });
    if (!responce.ok)
    {
        rnd = Math.floor(Math.random() * 3)+5;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">информация не записалась.</i>`
        checkbox.checked = !isAdmin;
        return;
    }
    else
    {
        rnd = Math.floor(Math.random() * 5)+3;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">информация записалась.</i>`
    }
}

async function deleteUser(userId) {
    item = document.getElementById(`user${userId}`);
    updateInfo = document.getElementById(`update-info`);
    data = {
        user_id : userId,
    };
    console.log(JSON.stringify(data));
    const responce = await fetch(`http://localhost:8080/museum${GLOBAL_museumId}/rights`,
    {
        method: "DELETE",
        credentials: 'include',
        body: JSON.stringify(data),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    });
    if (!responce.ok)
    {
        rnd = Math.floor(Math.random() * 3)+5;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">не удалось удалить пользователя.</i>`
        return;
    }
    else
    {
        item.outerHTML = ``;
        rnd = Math.floor(Math.random() * 5)+3;
        updateInfo.innerHTML = `<i style="color:#${rnd}${rnd}${rnd}">пользователь удалён.</i>`
    }
}

