import { useEffect } from "react";
import { FaBuilding, FaFacebook, FaPhone } from "react-icons/fa"
import { FaLocationPin } from "react-icons/fa6";
function About() {
	useEffect(()=>{
		window.scrollTo(0,0)
	},[])
    return (
        <div className="flex flex-col p-10 min-h-[60vh]">
			<div>
				<p className="text-3xl max-md:text-2xl font-semibold">About Us</p>
				<p>Realtime messaging platform is an online system that enable collective text-based commucation.</p> 
				<p>This system enables multiple people to commucate on a single platform/channel.</p>
				<p>All users are authenticated and authorzed to used this platform with their emails and passwords.</p> 
			</div>

			<div className="mt-10 text-sm">
				<p className="text-2xl text-slate-600 font-semibold">Contact Us</p>
				<p className="text-gray-600 mb-12">Discuss your interest with us.</p>
				<p className="text-xl text-slate-600 font-semibold">Contact Patnerships</p>
				<div className="flex flex-col gap-3 mt-3 pb-6 text-gray-600 border-b-[1px]">
					<div className="flex items-center gap-2">
						<FaBuilding/>
						<p>Rongo, Migori</p>
					</div>
					<div className="flex items-center gap-2">
						<FaLocationPin />
						<p>Rongo, Migori</p>
					</div>
					<div className="flex items-center gap-2">
						<FaPhone />
						<a href="tel:+254734720752" target="_blank" rel="noopener noreferrer">+254734720752</a>
					</div>
					<div className="flex items-center gap-2">
						<FaFacebook />
						<a href="https://facebook.com" target="_blank" rel="noopener noreferrer">@realtime_messaging</a>
					</div>
				</div>

				<div className="mt-6 pb-6 text-gray-600">
					<p className="text-xl text-slate-600 font-semibold">Contact Patnerships by Email</p>
					<p>If you wish to write us an email instead please use <a href="mailto:imranmat254@gmail.com" className="text-[var(--theme-blue)]" target="_blank" rel="noopener noreferrer">imranmat254@gmail.com</a></p>
				</div>
			</div>
        </div>
    );
};

export default About;
