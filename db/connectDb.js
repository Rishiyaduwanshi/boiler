import mongoose from "mongoose";
import { config } from '../config/index.js'

const connectDb = async () => {
  try {
    const connection = await mongoose.connect(config.MONGO_URI);
    console.log("Database connected successfully 🥳 on", connection.connection.host);
  } catch (error) {
    console.error("Database connection error 😢", error);
    process.exit(1);
  }
};

connectDb();

export default connectDb;

