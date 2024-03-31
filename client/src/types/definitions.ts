interface Event{
    id?:string,
    event:string,
    date:string,
    participants:string,
    remarks:string,
    location:string
}

interface Member{
    id?:string,
    name:string,
    telephone:number,
    email:string
}

interface Project{
    id?:string,
    project:string,
    member_contributions:number,
    account_for_payment:number
}

interface Song{
    id?:string,
    song_name:string,
    youtube_link:string
}

interface User{
    uid:string,
    photoURL:any,
    email:any,
    displayName:any,
    phoneNumber:any,
    emailVerified:boolean
}

interface Chat{
    chat_number:number,
    id?:string,
    from:string,
    message:string,
    time:string,
    today:string
}
export type{
    Event,
    Member,
    Project,
    Song,
    User,
    Chat,
}