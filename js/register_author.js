var GLOBAL_museumId = 0
window.onload=init;
async function init()
{    
    form = document.getElementById('form');
    form.addEventListener('submit', registerAuthor);
}

async function registerAuthor(event) 
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

    authorName = document.getElementById(`name`);
    birthYear = document.getElementById(`birth-year`);
    deathYear = document.getElementById(`death-year`);
    bio = document.getElementById(`biography`);
    if(authorName.value == '')
    {
        updateInfo.innerHTML = "Имя не указано"
        return
    }
    numBirthYear = parseInt(birthYear.value)
    numDeathYear = parseInt(deathYear.value)
    birthYearValid = (parseInt(birthYear.value)!=NaN)
    deathYearValid = (parseInt(deathYear.value)!=NaN)
    if((!birthYearValid && birthYear.value != "") || numBirthYear < 0)
    {
        updateInfo.innerHTML = "Год рождения указан некорректно"
        return
    }
    if((!deathYearValid && deathYear.value != "") || numDeathYear < 0)
    {
        updateInfo.innerHTML = "Год смерти указан некорректно"
        return
    }
    if (birthYearValid && deathYearValid && numBirthYear > numDeathYear)
    {
        updateInfo.innerHTML = "Год рождения превышает год смерти"
        return
    }
    data = {
        name: authorName.value,
        birth_year: {
            Int32: numBirthYear,
            Valid: birthYearValid,
        }, 
        death_year: {
            Int32: numDeathYear,
            Valid: deathYearValid,
        }, 
        biography: bio.value,
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