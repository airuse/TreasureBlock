// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

/**
 * @title GGGTX Token
 * @dev 标准ERC-20代币合约
 * @author Your Name
 * @notice 这是一个在Sepolia测试网络上部署的ERC-20代币
 */
contract GGGTX {
    // 代币基本信息
    string public constant name = "GGGTX Token";
    string public constant symbol = "GGGTX";
    uint8 public constant decimals = 18;
    
    // 代币总量
    uint256 public constant totalSupply = 1000000 * 10**decimals; // 1,000,000 GGGTX
    
    // 余额映射
    mapping(address => uint256) private _balances;
    
    // 授权映射
    mapping(address => mapping(address => uint256)) private _allowances;
    
    // 合约所有者
    address public owner;
    
    // 事件定义
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    // 修饰符
    modifier onlyOwner() {
        require(msg.sender == owner, "GGGTX: caller is not the owner");
        _;
    }
    
    /**
     * @dev 构造函数
     * @param _owner 合约所有者地址
     */
    constructor(address _owner) {
        require(_owner != address(0), "GGGTX: owner cannot be zero address");
        owner = _owner;
        _balances[_owner] = totalSupply;
        emit Transfer(address(0), _owner, totalSupply);
    }
    
    /**
     * @dev 查询账户余额
     * @param account 账户地址
     * @return 账户余额
     */
    function balanceOf(address account) public view returns (uint256) {
        return _balances[account];
    }
    
    /**
     * @dev 查询授权额度
     * @param account 账户地址
     * @param spender 被授权地址
     * @return 授权额度
     */
    function allowance(address account, address spender) public view returns (uint256) {
        return _allowances[account][spender];
    }
    
    /**
     * @dev 转账函数
     * @param to 接收地址
     * @param amount 转账数量
     * @return 是否成功
     */
    function transfer(address to, uint256 amount) public returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }
    
    /**
     * @dev 授权函数
     * @param spender 被授权地址
     * @param amount 授权数量
     * @return 是否成功
     */
    function approve(address spender, uint256 amount) public returns (bool) {
        _approve(msg.sender, spender, amount);
        return true;
    }
    
    /**
     * @dev 代理转账函数
     * @param from 发送地址
     * @param to 接收地址
     * @param amount 转账数量
     * @return 是否成功
     */
    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        uint256 currentAllowance = _allowances[from][msg.sender];
        require(currentAllowance >= amount, "GGGTX: transfer amount exceeds allowance");
        
        _transfer(from, to, amount);
        _approve(from, msg.sender, currentAllowance - amount);
        
        return true;
    }
    
    /**
     * @dev 增加授权额度
     * @param spender 被授权地址
     * @param addedValue 增加的授权数量
     * @return 是否成功
     */
    function increaseAllowance(address spender, uint256 addedValue) public returns (bool) {
        _approve(msg.sender, spender, _allowances[msg.sender][spender] + addedValue);
        return true;
    }
    
    /**
     * @dev 减少授权额度
     * @param spender 被授权地址
     * @param subtractedValue 减少的授权数量
     * @return 是否成功
     */
    function decreaseAllowance(address spender, uint256 subtractedValue) public returns (bool) {
        uint256 currentAllowance = _allowances[msg.sender][spender];
        require(currentAllowance >= subtractedValue, "GGGTX: decreased allowance below zero");
        _approve(msg.sender, spender, currentAllowance - subtractedValue);
        return true;
    }
    
    /**
     * @dev 内部转账函数
     * @param from 发送地址
     * @param to 接收地址
     * @param amount 转账数量
     */
    function _transfer(address from, address to, uint256 amount) internal {
        require(from != address(0), "GGGTX: transfer from the zero address");
        require(to != address(0), "GGGTX: transfer to the zero address");
        require(_balances[from] >= amount, "GGGTX: transfer amount exceeds balance");
        
        _balances[from] -= amount;
        _balances[to] += amount;
        
        emit Transfer(from, to, amount);
    }
    
    /**
     * @dev 内部授权函数
     * @param account 账户地址
     * @param spender 被授权地址
     * @param amount 授权数量
     */
    function _approve(address account, address spender, uint256 amount) internal {
        require(account != address(0), "GGGTX: approve from the zero address");
        require(spender != address(0), "GGGTX: approve to the zero address");
        
        _allowances[account][spender] = amount;
        emit Approval(account, spender, amount);
    }
    
    /**
     * @dev 转移合约所有权
     * @param newOwner 新所有者地址
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "GGGTX: new owner is the zero address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }
    
    /**
     * @dev 销毁代币（仅所有者可调用）
     * @param amount 销毁数量
     */
    function burn(uint256 amount) public onlyOwner {
        require(_balances[owner] >= amount, "GGGTX: burn amount exceeds balance");
        _balances[owner] -= amount;
        emit Transfer(owner, address(0), amount);
    }
    
    /**
     * @dev 批量转账（仅所有者可调用）
     * @param recipients 接收地址数组
     * @param amounts 转账数量数组
     */
    function batchTransfer(address[] memory recipients, uint256[] memory amounts) public onlyOwner {
        require(recipients.length == amounts.length, "GGGTX: arrays length mismatch");
        
        for (uint256 i = 0; i < recipients.length; i++) {
            _transfer(owner, recipients[i], amounts[i]);
        }
    }
}
