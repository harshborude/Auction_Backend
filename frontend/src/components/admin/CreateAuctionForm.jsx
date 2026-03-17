import { useState } from "react";
import styles from "./CreateAuctionForm.module.css";

function CreateAuctionForm({ onCreate }) {

    const [form, setForm] = useState({
        title: "",
        description: "",
        image_url: "",
        starting_price: "",
        bid_increment: "",
        end_time: ""
    });

    const handleChange = (e) => {
        setForm({ ...form, [e.target.name]: e.target.value });
    };

    const handleSubmit = () => {

        if (!form.title || !form.starting_price || !form.end_time) return;

        onCreate({
            Title: form.title,
            Description: form.description,
            ImageURL: form.image_url,
            StartingPrice: Number(form.starting_price),
            BidIncrement: Number(form.bid_increment || 1),
            EndTime: new Date(form.end_time).toISOString()
        });

        setForm({
            title: "",
            description: "",
            image_url: "",
            starting_price: "",
            bid_increment: "",
            end_time: ""
        });
    };

    return (
        <div className={styles.form}>

            <h3>Create Auction</h3>

            <input name="title" placeholder="Title" value={form.title} onChange={handleChange} />
            <input name="description" placeholder="Description" value={form.description} onChange={handleChange} />
            <input name="image_url" placeholder="Image URL" value={form.image_url} onChange={handleChange} />

            <input name="starting_price" type="number" placeholder="Starting Price" value={form.starting_price} onChange={handleChange} />
            <input name="bid_increment" type="number" placeholder="Bid Increment" value={form.bid_increment} onChange={handleChange} />

            <input name="end_time" type="datetime-local" value={form.end_time} onChange={handleChange} />

            <button onClick={handleSubmit}>Create</button>

        </div>
    );
}

export default CreateAuctionForm;