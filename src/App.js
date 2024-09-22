import React, { useState, useEffect } from 'react'
import { issueClass, transferAdmin, clearAdmin } from './api'
import { ToastContainer, toast } from 'react-toastify'
import 'react-toastify/dist/ReactToastify.css'
import { motion, AnimatePresence } from 'framer-motion'
import { Coins, UserPlus, UserMinus, ChevronRight } from 'lucide-react'
import coreumLogo from './assets/coreum-logo.png'
import './index.css'

export default function App() {
  const [activeSection, setActiveSection] = useState('issueClass')
  const [classData, setClassData] = useState({
    symbol: '',
    subunit: '',
    precision: 6,
    initial_amount: '',
    description: ''
  })
  const [adminData, setAdminData] = useState({
    denom: '',
    new_admin: ''
  })
  const [clearData, setClearData] = useState({
    denom: ''
  })

  const handleClassDataChange = (e) => setClassData({ ...classData, [e.target.name]: e.target.value })
  const handleAdminDataChange = (e) => setAdminData({ ...adminData, [e.target.name]: e.target.value })
  const handleClearDataChange = (e) => setClearData({ ...clearData, [e.target.name]: e.target.value })

  const handleIssueClass = async () => {
    try {
      const response = await issueClass(classData);
      // Extract the issuer address from the response
      const issuerAddress = response.issuer_address || response.sender || '';
      const fullDenom = `${classData.subunit}-${issuerAddress}`;
      console.log("Full denom:", fullDenom);
      localStorage.setItem('lastIssuedDenom', fullDenom);
      toast.success(`Transaction Successful! TxHash: ${response.transaction_id}`);
      toast.info(`Token issued with denom: ${fullDenom}`);
    } catch (error) {
      console.error("Error issuing class:", error);
      toast.error(`Error issuing class: ${error.response?.data || error.message}`);
    }
  };

  const handleTransferAdmin = async () => {
    try {
      const response = await transferAdmin(adminData)
      toast.success(`Admin Transferred! TxHash: ${response.transaction_id}`)
    } catch (error) {
      toast.error('Error transferring admin rights')
    }
  }

  const handleClearAdmin = async () => {
    try {
        console.log("Sending clearAdmin request with denom:", clearData.denom);

        if (!clearData.denom) {
            throw new Error("Please provide a valid denom to clear admin rights.");
        }

        const response = await clearAdmin(clearData);

        console.log("Clear admin response:", response);
        toast.success(`Admin Rights Cleared! TxHash: ${response.transaction_id}`);
    } catch (error) {
        console.error("Error clearing admin rights:", error.response || error);
        toast.error(`Error clearing admin rights: ${error.response?.data || error.message}`);
    }
};

  
  const renderSection = () => {
    switch (activeSection) {
      case 'issueClass':
        return (
          <FormSection
            title="Issue Token Class"
            icon={<Coins className="w-6 h-6 text-emerald-400" />}
            onSubmit={handleIssueClass}
            buttonText="Issue Class"
          >
            <Input
              placeholder="Class Symbol"
              name="symbol"
              value={classData.symbol}
              onChange={handleClassDataChange}
            />
            <Input
              placeholder="Subunit"
              name="subunit"
              value={classData.subunit}
              onChange={handleClassDataChange}
            />
            <Input
              type="number"
              placeholder="Precision"
              name="precision"
              value={classData.precision}
              onChange={handleClassDataChange}
            />
            <Input
              type="number"
              placeholder="Initial Amount"
              name="initial_amount"
              value={classData.initial_amount}
              onChange={handleClassDataChange}
            />
            <Input
              placeholder="Description"
              name="description"
              value={classData.description}
              onChange={handleClassDataChange}
            />
          </FormSection>
        )
      case 'transferAdmin':
        return (
          <FormSection
            title="Transfer Admin Rights"
            icon={<UserPlus className="w-6 h-6 text-emerald-400" />}
            onSubmit={handleTransferAdmin}
            buttonText="Transfer Admin"
          >
            <Input
              placeholder="Denom"
              name="denom"
              value={adminData.denom}
              onChange={handleAdminDataChange}
            />
            <Input
              placeholder="New Admin Address"
              name="new_admin"
              value={adminData.new_admin}
              onChange={handleAdminDataChange}
            />
          </FormSection>
        )
      case 'clearAdmin':
        return (
          <FormSection
            title="Clear Admin Rights"
            icon={<UserMinus className="w-6 h-6 text-emerald-400" />}
            onSubmit={handleClearAdmin}
            buttonText="Clear Admin"
          >
            <Input
              placeholder="Denom"
              name="denom"
              value={clearData.denom}
              onChange={handleClearDataChange}
            />
          </FormSection>
        )
      default:
        return null
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-900 via-gray-900 to-black flex flex-col items-center justify-center text-white font-sans p-4">
      <ToastContainer position="top-right" />
      <div className="w-full max-w-4xl flex flex-col items-center">
        <div className="w-full flex items-center justify-between mb-8">
          <img src={coreumLogo} alt="Coreum Logo" className="h-10" />
        </div>
        <h1 className="text-4xl md:text-5xl font-extrabold mb-10 text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-green-600">
          Fungible Token Management
        </h1>
        <div className="flex flex-wrap justify-center mb-10 gap-4">
          <TabButton
            active={activeSection === 'issueClass'}
            onClick={() => setActiveSection('issueClass')}
            icon={<Coins className="w-5 h-5" />}
          >
            Issue Class
          </TabButton>
          <TabButton
            active={activeSection === 'transferAdmin'}
            onClick={() => setActiveSection('transferAdmin')}
            icon={<UserPlus className="w-5 h-5" />}
          >
            Transfer Admin
          </TabButton>
          <TabButton
            active={activeSection === 'clearAdmin'}
            onClick={() => setActiveSection('clearAdmin')}
            icon={<UserMinus className="w-5 h-5" />}
          >
            Clear Admin
          </TabButton>
        </div>
        <AnimatePresence mode="wait">
          <motion.div
            key={activeSection}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className="w-full"
          >
            {renderSection()}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  )
}

function FormSection({ title, icon, children, onSubmit, buttonText }) {
  return (
    <div className="bg-gray-800 bg-opacity-50 backdrop-blur-lg shadow-xl rounded-xl p-8 w-full max-w-lg mx-auto">
      <h2 className="text-2xl font-bold mb-6 text-emerald-400 flex items-center">
        {icon}
        <span className="ml-2">{title}</span>
      </h2>
      <form onSubmit={(e) => { e.preventDefault(); onSubmit(); }}>
        {children}
        <Button type="submit">{buttonText}</Button>
      </form>
    </div>
  )
}

function Input({ ...props }) {
  return (
    <input
      {...props}
      className="mb-4 p-3 w-full bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-emerald-400 focus:border-transparent transition duration-200 ease-in-out"
    />
  )
}

function Button({ children, ...props }) {
  return (
    <button
      {...props}
      className="w-full bg-emerald-600 text-white py-3 rounded-lg font-semibold hover:bg-emerald-700 transition duration-200 ease-in-out transform hover:scale-105 flex items-center justify-center"
    >
      {children}
      <ChevronRight className="w-5 h-5 ml-2" />
    </button>
  )
}

function TabButton({ children, active, icon, ...props }) {
  return (
    <button
      {...props}
      className={`py-2 px-4 rounded-lg font-semibold flex items-center transition duration-200 ease-in-out ${
        active
          ? 'bg-emerald-600 text-white shadow-lg'
          : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
      }`}
    >
      {icon}
      <span className="ml-2">{children}</span>
    </button>
  )
}