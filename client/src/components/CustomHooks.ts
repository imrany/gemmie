import { toast } from 'react-hot-toast';

let useHotToast={
   error:(msg:string,opt?:any)=>{
      return toast.error(msg,opt)
   },
   sucess:(msg:string,opt?:any)=>{
      return toast.success(msg,opt)
   },
   info:(msg:string,opt?:any)=>{
      return toast.custom(msg,opt)
   }
}

export{
	useHotToast,
}
