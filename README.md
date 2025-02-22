# Идея проекта "Где находится картина?"
**Где находится картина** -- это веб-приложение, позволяющее отслеживать, в каком музее находится та или иная известная картина, а также выставляется ли она прямо сейчас.
## Описание сущностей и ER-модель
- **Картина.** Основной объект во всей структуре, так как цель данного проекта по сути накапливать данные о картинах. Необходимо создать гибкую систему описания картин, поэтому картины имеют уникальный идентификатор, а также могут иметь название, одного или нескольких авторов, год создания и прикрепленную оцифрованную копию. Картины могут создавать и редактировать только сотрудники институций (музеев), поэтому, чтобы редактировать информацию о той или иной картине необходимо либо быть сотрудником музея и иметь права на редактирование или самому создать эту картину (в рамках какого-то музея). Таким образом, картина непосредственно связана с институцией, которой она принадлежит. Кроме того, в случае, если картина в данный момент экспонируется, об этом указывается информация в отдельном поле.
- **Автор.** Автор картины. Необходимо хранить информацию об авторе, поскольку это может облегчить поиск для пользователя.
- **Пользователь.** Представляет собой простой аккаунт с логином и паролем. Для расширения этой функциональности добавляется возможность отмечать понравившиеся картины, чтобы следить за тем, экспонируются ли они где-либо. Пользователь системы может создать музей, тогда его аккаунт пользователя привяжется к этому музею правами редактировать картины в этом музее, а также давать другим пользователям права от имени этого музея.
- **Права.** Необходимы для хранения данных о том, что может делать пользователь в рамках того или иного музея. У одного пользователя могут быть права в рамках нескольких музеев.
- **Музей.** Записи о музеях создаются пользователями и содержат базовую информацию о них. Музей может быть подтверждённым. Это означает, что информация, предоставленная музеем веб-приложению скорее всего достоверна. Подтверждённость музея задается вручную администратором базы данных, поскольку полагается, что подтверждение музея требует тщательной проверки, выходящей за рамки онлайн-взаимодействия.
![WhereToFindThePainting drawio](https://github.com/user-attachments/assets/0bcb21a6-784b-404a-9c1d-abfe6fcc40bf)



