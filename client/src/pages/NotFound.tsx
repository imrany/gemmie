import { useNavigate, Link } from "react-router-dom";

function NotFound() {
    const navigate=useNavigate()
    return (
        <div className="flex flex-col h-[100vh] items-center justify-center">
            <p className="text-3xl font-semibold">404: Page Not Found</p>
            <p className="text-base my-2">Ooh..! This page is not available.</p>
            <div className="flex gap-10 mt-4">
                <Link to="/" className="button bg-[var(--theme-gray)]">Go Home</Link>
                <button className="button bg-[var(--theme-gray)]" onClick={()=>navigate(-1)}>Go Back</button>
            </div>
        </div>
    );
};

export default NotFound;