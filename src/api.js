import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

// Function to issue a class for an NFT or FT
export const issueClass = async(classData) => {
    try {
        const response = await axios.post(`${API_BASE_URL}/issue-token`, classData);
        console.log("Token issued with denom:", response.data.denom);
        // Store this denom for later use
        localStorage.setItem('lastIssuedDenom', response.data.denom);
        return response.data;
    } catch (error) {
        console.error("Error issuing class", error);
        throw error;
    }
};

// Function to transfer admin rights for a fungible token
export const transferAdmin = async(transferData) => {
    try {
        const response = await axios.post(`${API_BASE_URL}/transfer-admin`, transferData);
        return response.data;
    } catch (error) {
        console.error("Error transferring admin rights", error);
        throw error;
    }
};

export const clearAdmin = async (clearData) => {
    console.log("Clearing admin for denom:", clearData.denom);
    try {
        const response = await axios.post(`${API_BASE_URL}/clear-admin`, clearData); // Ensure correct URL
        return response.data;
    } catch (error) {
        console.error("Error in clearAdmin API call:", error.response || error);
        throw error;
    }
};
