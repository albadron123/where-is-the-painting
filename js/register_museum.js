var GLOBAL_museumId = 0
window.onload=init;
async function init()
{    
    form = document.getElementById('form');
    form.addEventListener('submit', registerMuseum);
}

async function registerMuseum(event) 
{
    event.preventDefault();

    updateInfo = document.getElementById(`update-info`);

    //sanity check

    /*
    type Museum struct {
	ID       int
	Name     string `gorm:"unique" binding:"required"`
	WebPage  string `binding:"required"`
	Verified bool   `gorm:"default:false"`
    }
    type Author struct {
        ID        int
        Name      string        `gorm:"not null"`
        BirthYear sql.NullInt32 `gorm:"check:(birth_year>=0)"`
        DeathYear sql.NullInt32 `gorm:"check:(death_year>birth_year)"`
        Biography string
    }
    */

    museumName = document.getElementById(`name`);
    webPage = document.getElementById(`web-page`);
    if(museumName.value == '')
    {
        updateInfo.innerHTML = "Имя не указано"
        return
    }
    if(webPage.value == '')
    {
        updateInfo.innerHTML = "Вебсайт не указан"
        return
    }
    data = {
        name: museumName.value,
        webpage: webPage.value,
    };
    console.log(JSON.stringify(data))
    const response = await fetch(window.location.href, {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify(data),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
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