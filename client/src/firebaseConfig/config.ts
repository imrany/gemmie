// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
import { getAuth, deleteUser, createUserWithEmailAndPassword, signInWithEmailAndPassword, onAuthStateChanged, signOut } from "firebase/auth";
import { getFirestore, collection, addDoc, getDocs, deleteDoc, doc, onSnapshot } from "firebase/firestore";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyDlkF6RhJVOYe5nyN6KguPH-F04m2ws1PE",
  authDomain: "realtime-messaging-platform.firebaseapp.com",
  projectId: "realtime-messaging-platform",
  storageBucket: "realtime-messaging-platform.appspot.com",
  messagingSenderId: "368601117784",
  appId: "1:368601117784:web:9da71600bce858350dc62d",
  measurementId: "G-35BFDWK8PF"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const auth = getAuth(app);
const db= getFirestore(app)

export{
  auth,
  db,
  collection,
  addDoc,
  getDocs,
  deleteDoc,
  onSnapshot,
  doc,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
  deleteUser,
}
